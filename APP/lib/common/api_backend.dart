import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';
import 'package:file_picker/file_picker.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:http/http.dart' as http;
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:ogree_app/models/container.dart';
import 'package:ogree_app/models/domain.dart';
import 'package:ogree_app/models/netbox.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/models/user.dart';
import 'package:universal_html/html.dart';

import 'definitions.dart';

part 'api_tenant.dart';

// Globals
String apiUrl = "";
var token = "";
String tenantName = "";
bool isTenantAdmin = false; // a tenant admin can access its config page
String tenantUrl = ""; // used by SuperAdmin to connect between tenant APIs
var tenantToken = ""; // used by SuperAdmin to connect between tenant APIs
BackendType backendType = BackendType.tenant;

enum BackendType { docker, kubernetes, tenant, unavailable }

// Helper Functions
getHeader(token) => {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Authorization': 'Bearer $token',
    };

String reformatDate(String date) {
  // dd/MM/yyyy -> yyyy-MM-dd
  List<String> dateParts = date.split("/");
  return "${dateParts[2]}-${dateParts[1]}-${dateParts[0]}";
}

String urlDateAppend(String dateRange) {
  var ranges = dateRange.split(" - ");
  String urlAppend = "&startDate=${reformatDate(ranges[0])}";
  if (ranges.length > 1) {
    urlAppend = "$urlAppend&endDate=${reformatDate(ranges[1])}";
  }
  return urlAppend;
}

// API calls
Future<Result<List<String>, Exception>> loginAPI(String email, String password,
    {String userUrl = ""}) async {
  // Make sure it is clean
  tenantUrl = "";
  isTenantAdmin = false;
  token = "";
  tenantToken = "";

  // Set request
  if (userUrl != "") {
    apiUrl = userUrl;
  } else {
    apiUrl = dotenv.get('API_URL', fallback: 'http://localhost:3001');
  }
  print("API login ogree $apiUrl");
  try {
    Uri url = Uri.parse('$apiUrl/api/login');
    final response = await http.post(url,
        body: json
            .encode(<String, String>{'email': email, 'password': password}));

    // Handle response
    Map<String, dynamic> data = json.decode(response.body);
    if (response.statusCode == 200) {
      data = (Map<String, dynamic>.from(data["account"]));
      token = data["token"]!;
      if (data["isTenant"] == null && data["roles"]["*"] == "manager") {
        // Not tenant mode, but tenant admin
        isTenantAdmin = true;
        tenantUrl = apiUrl;
        tenantToken = token;
      }
      if (data["isKubernetes"] == true) {
        // is Kubernetes API
        backendType = BackendType.kubernetes;
      }
      return Success([data["email"].toString(), data["isTenant"] ?? ""]);
    } else {
      return Failure(Exception());
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<BackendType, Exception>> fetchApiVersion(String urlApi,
    {http.Client? client}) async {
  print("API get TenantName");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$urlApi/api/version');
    final response = await client.get(url, headers: getHeader(token));
    print(response.statusCode);
    if (response.statusCode == 200) {
      Map<String, dynamic> data = json.decode(response.body);
      if (data["isKubernetes"] != null) {
        if (data["isKubernetes"] == true) {
          backendType = BackendType.kubernetes;
          return Success(BackendType.kubernetes);
        } else {
          backendType = BackendType.docker;
          return Success(BackendType.docker);
        }
      } else {
        data = (Map<String, dynamic>.from(data["data"]));
        if (data.isNotEmpty || data["Customer"] != null) {
          tenantName = data["Customer"];
          backendType = BackendType.tenant;
          print(tenantName);
          return Success(BackendType.tenant);
        } else {
          backendType = BackendType.unavailable;
          return Success(BackendType.unavailable);
        }
      }
    } else if (response.statusCode == 403) {
      backendType = BackendType.tenant;
      return Success(BackendType.tenant);
    } else {
      return Failure(Exception("Unable to get version from server"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> changeUserPassword(
    String currentPassword, newPassword) async {
  print("API change password");
  try {
    Uri url = Uri.parse('$apiUrl/api/users/password/change');
    final response = await http.post(url,
        body: json.encode(<String, dynamic>{
          'currentPassword': currentPassword,
          'newPassword': newPassword
        }),
        headers: getHeader(token));
    print(response.statusCode);
    Map<String, dynamic> data = json.decode(response.body);
    if (response.statusCode == 200) {
      token = data["token"]!;
      return const Success(null);
    } else {
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> userForgotPassword(String email,
    {String userUrl = ""}) async {
  print("API forgot password");
  if (userUrl != "") {
    apiUrl = userUrl;
  } else {
    apiUrl = dotenv.get('API_URL', fallback: 'http://localhost:3001');
  }
  try {
    Uri url = Uri.parse('$apiUrl/api/users/password/forgot');
    final response = await http.post(
      url,
      body: json.encode(<String, dynamic>{'email': email}),
    );
    print(response.statusCode);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> userResetPassword(
    String password, String resetToken,
    {String userUrl = ""}) async {
  print("API reset password");
  if (userUrl != "") {
    apiUrl = userUrl;
  } else {
    apiUrl = dotenv.get('API_URL', fallback: 'http://localhost:3001');
  }
  try {
    Uri url = Uri.parse('$apiUrl/api/users/password/reset');
    final response = await http.post(
      url,
      body: json.encode(<String, dynamic>{'newPassword': password}),
      headers: getHeader(resetToken),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<Map<String, List<String>>>, Exception>> fetchObjectsTree(
    {Namespace namespace = Namespace.Physical,
    String dateRange = "",
    bool isTenantMode = false}) async {
  print("API get tree: NS=$namespace");

  // Define URL and token to use
  String localUrl = '/api/hierarchy';
  String localToken = token;
  if (isTenantMode) {
    localUrl = tenantUrl + localUrl;
    localToken = tenantToken;
  } else {
    localUrl = apiUrl + localUrl;
  }
  // Add filters, if any
  String namespaceStr = namespace.name.toLowerCase();
  if (namespace == Namespace.Physical) {
    localUrl = '$localUrl?namespace=$namespaceStr&withcategories=true';
  } else {
    localUrl = '$localUrl?namespace=$namespaceStr';
  }
  if (dateRange != "") {
    localUrl = localUrl + urlDateAppend(dateRange);
  }

  // Request
  try {
    Uri url = Uri.parse(localUrl);
    final response = await http.get(url, headers: getHeader(localToken));
    print(response.statusCode);
    if (response.statusCode == 200) {
      // Convert dynamic Map to expected type
      Map<String, dynamic> data = json.decode(response.body);
      data = (Map<String, dynamic>.from(data["data"]));
      Map<String, Map<String, dynamic>> converted = {};
      Map<String, Map<String, dynamic>> converted2 = {};
      Map<String, List<String>> tree = {};
      Map<String, List<String>> categories = {};
      for (var item in data.keys) {
        converted[item.toString()] = Map<String, dynamic>.from(data[item]);
      }
      for (var item in converted["tree"]!.keys) {
        converted2[item.toString()] =
            Map<String, dynamic>.from(converted["tree"]![item]!);
      }
      for (var item in converted2[namespaceStr]!.keys) {
        tree[item.toString()] =
            List<String>.from(converted2[namespaceStr]![item]);
      }
      // Namespace adaptations
      if (namespace == Namespace.Physical) {
        tree["*"]!.addAll(tree["*stray_object"]!);
        for (var item in converted["categories"]!.keys) {
          categories[item.toString()] =
              List<String>.from(converted["categories"]![item]);
        }
      } else if (namespace == Namespace.Logical) {
        tree["*"] = tree.keys.toList();
      }
      return Success([tree, categories]);
    } else {
      return Failure(
          Exception('${response.statusCode}: Failed to load objects'));
    }
  } on Exception catch (e) {
    print(e.toString());
    return Failure(e);
  }
}

Future<Result<Map<String, Map<String, String>>, Exception>>
    fetchAttributes() async {
  print("API get Attrs");
  try {
    Uri url = Uri.parse('$apiUrl/api/hierarchy/attributes');
    final response = await http.get(url, headers: getHeader(token));
    print(response.statusCode);
    if (response.statusCode == 200) {
      // Convert dynamic Map to expected type
      Map<String, dynamic> data = json.decode(response.body);
      data = (Map<String, dynamic>.from(data["data"]));
      Map<String, Map<String, String>> converted = {};
      for (var item in data.keys) {
        converted[item.toString()] = Map<String, String>.from(data[item]);
      }
      return Success(converted);
    } else {
      return Failure(
          Exception('${response.statusCode}: Failed to load objects'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<Project>, Exception>> fetchProjects(String userEmail,
    {http.Client? client}) async {
  print("API get Projects");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$apiUrl/api/projects?user=$userEmail');
    final response = await client.get(url, headers: getHeader(token));
    print(response.statusCode);
    if (response.statusCode == 200) {
      print(response);
      print(response.body);
      // Convert dynamic Map to expected type
      Map<String, dynamic> data = json.decode(response.body);
      data = (Map<String, dynamic>.from(data["data"]));
      List<Project> projects = [];
      for (var project in data["projects"]) {
        projects.add(Project.fromMap(project));
      }
      return Success(projects);
    } else {
      return Failure(
          Exception('${response.statusCode}: Failed to load objects'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> deleteProject(String id) async {
  print("API delete Projects");
  try {
    Uri url = Uri.parse('$apiUrl/api/projects/$id');
    final response = await http.delete(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> modifyProject(Project project) async {
  print("API modify Projects");
  try {
    Uri url = Uri.parse('$apiUrl/api/projects/${project.id}');
    final response =
        await http.put(url, body: project.toJson(), headers: getHeader(token));
    print(response);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createProject(Project project) async {
  print("API create Projects");
  try {
    Uri url = Uri.parse('$apiUrl/api/projects');
    final response =
        await http.post(url, body: project.toJson(), headers: getHeader(token));
    print(response);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<(List<Tenant>, List<DockerContainer>), Exception>>
    fetchApplications({http.Client? client}) async {
  print("API get Apps");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$apiUrl/api/apps');
    final response = await client.get(url, headers: getHeader(token));
    print(response.statusCode);
    if (response.statusCode == 200) {
      Map<String, dynamic> data = json.decode(response.body);
      List<Tenant> tenants = [];
      for (var project in data["tenants"]) {
        tenants.add(Tenant.fromMap(project));
      }
      List<DockerContainer> containers = [];
      for (var tool in data["tools"]) {
        var container = DockerContainer.fromMap(tool);
        if (container.ports.isNotEmpty) {
          container.ports = "http://${container.ports.split("-").first}";
          container.ports =
              container.ports.replaceFirst("0.0.0.0", "localhost");
        }
        containers.add(container);
      }
      return Success((tenants, containers));
    } else {
      return Failure(
          Exception('${response.statusCode}: Failed to load objects'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Stream<String>, Exception>> createTenant(Tenant tenant) async {
  print("API create Tenants");
  try {
    return connectStream('POST', '$apiUrl/api/tenants', tenant.toJson());
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Stream<String>, Exception>> updateTenant(Tenant tenant) async {
  print("API update Tenants");
  try {
    return connectStream(
        'PUT', '$apiUrl/api/tenants/${tenant.name}', tenant.toJson());
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Stream<String>, Exception>> connectStream(
    String method, String urlStr, String body) async {
  if (kIsWeb) {
    // Special SSE handling for web
    int progress = 0;
    final httpRequest = HttpRequest();
    final streamController = StreamController<String>();
    httpRequest.open(method, urlStr);
    getHeader(token).forEach((key, value) {
      httpRequest.setRequestHeader(key, value);
    });
    httpRequest.onProgress.listen((event) {
      final data = httpRequest.responseText!.substring(progress);
      progress += data.length;
      streamController.add(data);
    });
    httpRequest.addEventListener('loadend', (event) {
      httpRequest.abort();
      streamController.close();
    });
    httpRequest.addEventListener('error', (event) {
      streamController.add(
        'Error in backend connection',
      );
    });
    httpRequest.send(body);
    return Success(streamController.stream);
  } else {
    // SSE handle for other builds
    Uri url = Uri.parse(urlStr);
    final client = http.Client();
    var request = http.Request(method, url)..headers.addAll(getHeader(token));
    request.body = body;
    final response = await client.send(request);
    if (response.statusCode == 200) {
      return Success(response.stream.toStringStream());
    } else {
      return Failure(
          Exception("Error processing tenant ${response.statusCode}"));
    }
  }
}

Future<Result<void, Exception>> uploadImage(
    PlatformFile image, String tenant) async {
  print("API upload Tenant logo");
  try {
    Uri url = Uri.parse('$apiUrl/api/tenants/$tenant/logo');
    var request = http.MultipartRequest("POST", url);
    request.headers.addAll(getHeader(token));
    request.files.add(http.MultipartFile.fromBytes("file", image.bytes!,
        filename: image.name));

    var response = await request.send();
    print(response.statusCode);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      String errorMsg = await response.stream.bytesToString();
      return Failure(Exception(errorMsg));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<dynamic, Exception>> backupTenantDB(
    String tenantName, String password, bool shouldDownload) async {
  print("API backup Tenants");
  try {
    Uri url = Uri.parse('$apiUrl/api/tenants/$tenantName/backup');
    final response = await http.post(url,
        body: json.encode(<String, dynamic>{
          'password': password,
          'shouldDownload': shouldDownload,
        }),
        headers: getHeader(token));
    print(response);
    if (response.statusCode == 200) {
      if (shouldDownload) {
        return Success(response.bodyBytes);
      } else {
        return Success(response.body.toString());
      }
    } else {
      String data = json.decode(response.body);
      return Failure(Exception("Error backing up tenant $data"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createBackendServer(
    Map<String, dynamic> newBackend) async {
  print("API create Back Server");
  try {
    Uri url = Uri.parse('$apiUrl/api/servers');
    final response = await http.post(url,
        body: json.encode(newBackend), headers: getHeader(token));
    print(response);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      return Failure(Exception("Error creating backend: ${response.body}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> deleteTenant(String objName,
    {http.Client? client}) async {
  print("API delete Tenant");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$apiUrl/api/tenants/$objName');
    final response = await client.delete(url, headers: getHeader(token));
    print(response.statusCode);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      return Failure(Exception("Error deleting tenant: ${response.body}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<DockerContainer>, Exception>> fetchTenantDockerInfo(
    String tenantName,
    {http.Client? client}) async {
  print("API get Tenant Docker Info");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$apiUrl/api/tenants/${tenantName.toLowerCase()}');
    final response = await client.get(url, headers: getHeader(token));
    print(response.statusCode);
    if (response.statusCode == 200) {
      List<dynamic> data = json.decode(response.body);
      List<DockerContainer> converted = [];
      for (var item in data) {
        converted.add(DockerContainer.fromMap(item));
      }
      return Success(converted);
    } else {
      print('${response.statusCode}: ${response.body}');
      return Failure(Exception('${response.statusCode}: ${response.body}'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<String, Exception>> fetchContainerLogs(String name,
    {http.Client? client}) async {
  print("API get Container Logs $name");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$apiUrl/api/containers/$name');
    final response = await client.get(url, headers: getHeader(token));
    print(response.statusCode);
    if (response.statusCode == 200) {
      Map<String, dynamic> data = json.decode(response.body);
      return Success(data["logs"].toString());
    } else {
      return Failure(Exception('${response.statusCode}: failed to load logs'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createNetbox(Netbox netbox) async {
  print("API create Netbox");
  try {
    Uri url = Uri.parse('$apiUrl/api/tools/netbox');
    final response =
        await http.post(url, body: netbox.toJson(), headers: getHeader(token));
    print(response);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      String data = json.decode(response.body);
      return Failure(Exception("Error creating netbox $data"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createOpenDcim(
    String dcimPort, adminerPort) async {
  print("API create OpenDCIM");
  try {
    Uri url = Uri.parse('$apiUrl/api/tools/opendcim');
    final response = await http.post(url,
        body: json.encode(<String, dynamic>{
          'dcimPort': dcimPort,
          'adminerPort': adminerPort,
        }),
        headers: getHeader(token));
    print(response);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      String data = json.decode(response.body);
      return Failure(Exception("Error creating netbox $data"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> deleteTool(String tool) async {
  print("API delete Netbox");
  try {
    Uri url = Uri.parse('$apiUrl/api/tools/$tool');
    final response = await http.delete(url, headers: getHeader(token));
    print(response);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      String data = json.decode(response.body);
      return Failure(Exception("Error deleting netbox $data"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> uploadNetboxDump(PlatformFile file) async {
  print("API upload netbox dump");
  try {
    Uri url = Uri.parse('$apiUrl/api/tools/netbox/dump');
    var request = http.MultipartRequest("POST", url);
    request.headers.addAll(getHeader(token));
    request.files.add(
        http.MultipartFile.fromBytes("file", file.bytes!, filename: file.name));
    var response = await request.send();
    print(response.statusCode);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      String errorMsg = await response.stream.bytesToString();
      return Failure(Exception(errorMsg));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> importNetboxDump() async {
  print("API import dump Netbox");
  try {
    Uri url = Uri.parse('$apiUrl/api/tools/netbox/import');
    final response = await http.post(url, headers: getHeader(token));
    print(response);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      String data = json.decode(response.body);
      return Failure(Exception("Error importing netbox dump: $data"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}
