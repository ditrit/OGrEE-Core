import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:http/http.dart' as http;
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/models/alert.dart';
import 'package:ogree_app/models/container.dart';
import 'package:ogree_app/models/domain.dart';
import 'package:ogree_app/models/netbox.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/models/tag.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/models/user.dart';
import 'package:universal_html/html.dart';

part 'api_tenant.dart';

// Globals
String apiUrl = "";
String token = "";
String tenantName = "";
bool isTenantAdmin = false; // a tenant admin can access its config page
String tenantUrl = ""; // used by SuperAdmin to connect between tenant APIs
String tenantToken = ""; // used by SuperAdmin to connect between tenant APIs
BackendType backendType = BackendType.tenant;

enum BackendType { docker, kubernetes, tenant, unavailable }

// Helper Functions
Map<String, String> getHeader(token) => {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Authorization': 'Bearer $token',
    };

String reformatDate(String date) {
  // dd/MM/yyyy -> yyyy-MM-dd
  final List<String> dateParts = date.split("/");
  return "${dateParts[2]}-${dateParts[1]}-${dateParts[0]}";
}

String urlDateAppend(String dateRange) {
  final ranges = dateRange.split(" - ");
  String urlAppend = "&startDate=${reformatDate(ranges[0])}";
  if (ranges.length > 1) {
    urlAppend = "$urlAppend&endDate=${reformatDate(ranges[1])}";
  }
  return urlAppend;
}

String wrapResponseMsg(http.Response response, {String? message}) {
  if (response.statusCode == 401) {
    return "Votre session a expiré, veuillez vous reconnecter pour exécuter cette action";
  }
  return '${response.statusCode}: ${message ?? response.reasonPhrase}';
}

// API calls
Future<Result<List<String>, Exception>> loginAPI(
  String email,
  String password, {
  String userUrl = "",
  bool stayLoggedIn = false,
}) async {
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
  try {
    final Uri url = Uri.parse('$apiUrl/api/login');
    final response = await http.post(
      url,
      body: json.encode(<String, dynamic>{
        'email': email,
        'password': password,
        'stayLoggedIn': stayLoggedIn,
      }),
    );

    // Handle response
    Map<String, dynamic> data = json.decode(response.body);
    if (response.statusCode == 200) {
      data = Map<String, dynamic>.from(data["account"]);
      token = data["token"]!;
      if (data["isTenant"] == null &&
          data["roles"][allDomainsTag] == "manager") {
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

Future<Result<BackendType, Exception>> fetchApiVersion(
  String urlApi, {
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$urlApi/api/version');
    final response = await client.get(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      Map<String, dynamic> data = json.decode(response.body);
      if (data["isKubernetes"] != null) {
        if (data["isKubernetes"] == true) {
          backendType = BackendType.kubernetes;
          return const Success(BackendType.kubernetes);
        } else {
          backendType = BackendType.docker;
          return const Success(BackendType.docker);
        }
      } else {
        data = Map<String, dynamic>.from(data["data"]);
        if (data.isNotEmpty || data["Customer"] != null) {
          tenantName = data["Customer"];
          backendType = BackendType.tenant;
          return const Success(BackendType.tenant);
        } else {
          backendType = BackendType.unavailable;
          return const Success(BackendType.unavailable);
        }
      }
    } else if (response.statusCode == 403) {
      backendType = BackendType.tenant;
      return const Success(BackendType.tenant);
    } else {
      return Failure(Exception("Unable to get version from server"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> changeUserPassword(
  String currentPassword,
  newPassword,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/users/password/change');
    final response = await http.post(
      url,
      body: json.encode(<String, dynamic>{
        'currentPassword': currentPassword,
        'newPassword': newPassword,
      }),
      headers: getHeader(token),
    );
    final Map<String, dynamic> data = json.decode(response.body);
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

Future<Result<void, Exception>> userForgotPassword(
  String email, {
  String userUrl = "",
}) async {
  if (userUrl != "") {
    apiUrl = userUrl;
  } else {
    apiUrl = dotenv.get('API_URL', fallback: 'http://localhost:3001');
  }
  try {
    final Uri url = Uri.parse('$apiUrl/api/users/password/forgot');
    final response = await http.post(
      url,
      body: json.encode(<String, dynamic>{'email': email}),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> userResetPassword(
  String password,
  String resetToken, {
  String userUrl = "",
}) async {
  if (userUrl != "") {
    apiUrl = userUrl;
  } else {
    apiUrl = dotenv.get('API_URL', fallback: 'http://localhost:3001');
  }
  try {
    final Uri url = Uri.parse('$apiUrl/api/users/password/reset');
    final response = await http.post(
      url,
      body: json.encode(<String, dynamic>{'newPassword': password}),
      headers: getHeader(resetToken),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<Map<String, List<String>>>, Exception>> fetchObjectsTree({
  Namespace namespace = Namespace.Physical,
  String dateRange = "",
  bool isTenantMode = false,
}) async {
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
  final String namespaceStr = namespace.name.toLowerCase();
  if (namespace == Namespace.Physical || namespace == Namespace.Logical) {
    localUrl = '$localUrl?namespace=$namespaceStr&withcategories=true';
  } else {
    localUrl = '$localUrl?namespace=$namespaceStr';
  }
  if (dateRange != "") {
    localUrl = localUrl + urlDateAppend(dateRange);
  }

  // Request
  try {
    final Uri url = Uri.parse(localUrl);
    final response = await http.get(url, headers: getHeader(localToken));
    if (response.statusCode == 200) {
      // Convert dynamic Map to expected type
      Map<String, dynamic> data = json.decode(response.body);
      data = Map<String, dynamic>.from(data["data"]);
      final Map<String, Map<String, dynamic>> converted = {};
      final Map<String, Map<String, dynamic>> converted2 = {};
      final Map<String, List<String>> tree = {};
      final Map<String, List<String>> categories = {};
      for (final item in data.keys) {
        converted[item] = Map<String, dynamic>.from(data[item]);
      }
      for (final item in converted["tree"]!.keys) {
        converted2[item] = Map<String, dynamic>.from(converted["tree"]![item]!);
      }
      for (final item in converted2[namespaceStr]!.keys) {
        tree[item] = List<String>.from(converted2[namespaceStr]![item]);
      }
      // Namespace adaptations
      if (namespace == Namespace.Physical || namespace == Namespace.Logical) {
        for (final item in converted["categories"]!.keys) {
          categories[item] = List<String>.from(converted["categories"]![item]);
        }
        if (namespace == Namespace.Physical) {
          if (tree["*stray_object"] != null) {
            tree["*"]!.addAll(tree["*stray_object"]!);
          }
        } else if (namespace == Namespace.Logical) {
          tree["*"] = tree.keys.where((e) => e.contains("*")).toList();
        }
      }
      return Success([tree, categories]);
    } else {
      return Failure(
        Exception('${response.statusCode}: Failed to load objects'),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Map<String, Map<String, dynamic>>, Exception>>
    fetchAttributes() async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/hierarchy/attributes');
    final response = await http.get(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      // Convert dynamic Map to expected type
      Map<String, dynamic> data = json.decode(response.body);
      data = Map<String, dynamic>.from(data["data"]);
      final Map<String, Map<String, dynamic>> converted = {};
      for (final item in data.keys) {
        converted[item] = Map<String, dynamic>.from(data[item]);
      }
      return Success(converted);
    } else {
      return Failure(
        Exception('${response.statusCode}: Failed to load objects'),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<Project>, Exception>> fetchProjects(
  String userEmail, {
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$apiUrl/api/projects?user=$userEmail');
    final response = await client.get(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      // Convert dynamic Map to expected type
      Map<String, dynamic> data = json.decode(response.body);
      data = Map<String, dynamic>.from(data["data"]);
      final List<Project> projects = [];
      for (final project in data["projects"]) {
        projects.add(Project.fromMap(project));
      }
      return Success(projects);
    } else {
      return Failure(
        Exception(
          wrapResponseMsg(response, message: 'Failed to load objects'),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> modifyProject(Project project) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/projects/${project.id}');
    final response =
        await http.put(url, body: project.toJson(), headers: getHeader(token));
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
  try {
    final Uri url = Uri.parse('$apiUrl/api/projects');
    final response =
        await http.post(url, body: project.toJson(), headers: getHeader(token));
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

Future<Result<void, Exception>> createAlert(Alert alert) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/alerts');
    final response =
        await http.post(url, body: alert.toJson(), headers: getHeader(token));
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

Future<Result<List<Alert>, Exception>> fetchAlerts({
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$apiUrl/api/alerts');
    final response = await client.get(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      // Convert dynamic Map to expected type
      Map<String, dynamic> data = json.decode(response.body);
      data = Map<String, dynamic>.from(data["data"]);
      final List<Alert> alerts = [];
      for (final alert in data["alerts"]) {
        alerts.add(Alert.fromMap(alert));
      }
      return Success(alerts);
    } else {
      return Failure(
        Exception(
          wrapResponseMsg(response, message: 'Failed to load objects'),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Alert, Exception>> fetchAlert(
  String id, {
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$apiUrl/api/alerts/$id');
    final response = await client.get(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      // Convert dynamic Map to expected type
      Map<String, dynamic> data = json.decode(response.body);
      data = Map<String, dynamic>.from(data["data"]);
      return Success(Alert.fromMap(data));
    } else {
      return Failure(
        Exception(wrapResponseMsg(response, message: 'Failed to get alert')),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createObject(
  Map<String, dynamic> object,
  String category,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/${category}s');
    final response = await http.post(
      url,
      body: json.encode(object),
      headers: getHeader(token),
    );
    if (response.statusCode == 200 || response.statusCode == 201) {
      return const Success(null);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      if (data["errors"] != null) {
        final errors = List<String>.from(data["errors"]);
        String errStr = "";
        for (final err in errors) {
          errStr = "$errStr\n$err";
        }
        return Failure(Exception(errStr));
      }
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Map<String, dynamic>, Exception>> fetchObject(
  String id,
  AppLocalizations localeMsg, {
  String idKey = "id",
}) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/objects?$idKey=$id');
    final response = await http.get(url, headers: getHeader(token));
    if (response.statusCode == 200 || response.statusCode == 201) {
      final Map<String, dynamic> data = json.decode(response.body);
      final list = List<Map<String, dynamic>>.from(data["data"]);
      if (list.isEmpty) {
        return Failure(Exception(localeMsg.noObjectsFound));
      }
      return Success(list.first);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Map<String, dynamic>, Exception>> fetchObjectChildren(
  String id,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/hierarchy_objects/$id/all?limit=2');
    final response = await http.get(url, headers: getHeader(token));
    if (response.statusCode == 200 || response.statusCode == 201) {
      final Map<String, dynamic> data = json.decode(response.body);
      return Success(Map<String, dynamic>.from(data["data"]));
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Map<String, dynamic>, Exception>> fetchObjectImpact(
  String id,
  List<String> categories,
  List<String> ptypes,
  List<String> vtypes,
) async {
  String queryParam = listToQueryParam("", categories, "categories");
  queryParam = listToQueryParam(queryParam, ptypes, "ptypes");
  queryParam = listToQueryParam(queryParam, vtypes, "vtypes");
  try {
    String urlStr = '$apiUrl/api/impact/$id';
    if (queryParam.isNotEmpty) {
      urlStr = "$urlStr?$queryParam";
    }
    final Uri url = Uri.parse(urlStr);
    final response = await http.get(url, headers: getHeader(token));
    if (response.statusCode == 200 || response.statusCode == 201) {
      final Map<String, dynamic> data = json.decode(response.body);
      return Success(Map<String, dynamic>.from(data["data"]));
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

String listToQueryParam(String currentParam, List<String> list, String key) {
  String param = currentParam;
  for (final String item in list) {
    if (param.isNotEmpty) {
      param = "$param&";
    }
    param = "$param$key=$item";
  }
  return param;
}

Future<Result<List<Map<String, dynamic>>, Exception>> fetchWithComplexFilter(
  String filter,
  Namespace namespace,
  AppLocalizations localeMsg,
) async {
  try {
    final Uri url = Uri.parse(
      '$apiUrl/api/objects/search?namespace=${namespace.name.toLowerCase()}',
    );
    final response = await http.post(
      url,
      body: json.encode(<String, dynamic>{'filter': filter}),
      headers: getHeader(token),
    );
    if (response.statusCode == 200 || response.statusCode == 201) {
      final Map<String, dynamic> data = json.decode(response.body);
      final list = List<Map<String, dynamic>>.from(data["data"]);
      if (list.isEmpty) {
        return Failure(Exception(localeMsg.noObjectsFound));
      }
      return Success(list);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> updateObject(
  String objId,
  String category,
  Map<String, dynamic> object,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/${category}s/$objId');
    final response = await http.put(
      url,
      body: json.encode(object),
      headers: getHeader(token),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final data = json.decode(response.body);
      if (data["errors"] != null) {
        final errors = List<String>.from(data["errors"]);
        String errStr = "";
        for (final err in errors) {
          errStr = "$errStr\n$err";
        }
        return Failure(Exception(errStr));
      }
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> deleteObject(
  String objId,
  String category, {
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$apiUrl/api/${category}s/$objId');
    final response = await client.delete(url, headers: getHeader(token));
    if (response.statusCode >= 200 && response.statusCode < 300) {
      return const Success(null);
    } else {
      final data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createTemplate(
  Uint8List file,
  String category,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/${category}s');
    final response =
        await http.post(url, body: file, headers: getHeader(token));
    if (response.statusCode == 200 || response.statusCode == 201) {
      return const Success(null);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<String>, Exception>> fetchGroupContent(
  String id,
  category,
  AppLocalizations localeMsg,
) async {
  try {
    final Uri url =
        Uri.parse('$apiUrl/api/objects?id=$id.*&category=$category');
    final response = await http.get(url, headers: getHeader(token));
    if (response.statusCode == 200 || response.statusCode == 201) {
      final Map<String, dynamic> data = json.decode(response.body);
      final list = List<Map<String, dynamic>>.from(data["data"]);
      if (list.isEmpty) {
        return Failure(Exception(localeMsg.noObjectsFound));
      } else {
        final List<String> content = [];
        for (final item in list) {
          content.add(item["name"].toString());
        }
        return Success(content);
      }
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
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$apiUrl/api/apps');
    final response = await client.get(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      final Map<String, dynamic> data = json.decode(response.body);
      final List<Tenant> tenants = [];
      for (final project in data["tenants"]) {
        tenants.add(Tenant.fromMap(project));
      }
      final List<DockerContainer> containers = [];
      for (final tool in data["tools"]) {
        final container = DockerContainer.fromMap(tool);
        if (container.ports.isNotEmpty) {
          container.ports =
              "http://${container.ports.split(",").last.split("-").first.trim()}";
          container.ports =
              container.ports.replaceFirst("0.0.0.0", "localhost");
        }
        containers.add(container);
      }
      return Success((tenants, containers));
    } else {
      return Failure(
        Exception(
          wrapResponseMsg(response, message: 'Failed to load objects'),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Stream<String>, Exception>> createTenant(Tenant tenant) async {
  try {
    return connectStream('POST', '$apiUrl/api/tenants', tenant.toJson());
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Stream<String>, Exception>> updateTenant(Tenant tenant) async {
  try {
    return connectStream(
      'PUT',
      '$apiUrl/api/tenants/${tenant.name}',
      tenant.toJson(),
    );
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Stream<String>, Exception>> stopTenant(String tenantName) async {
  try {
    return connectStream('POST', '$apiUrl/api/tenants/$tenantName/stop', "");
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Stream<String>, Exception>> startTenant(String tenantName) async {
  try {
    return connectStream('POST', '$apiUrl/api/tenants/$tenantName/start', "");
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Stream<String>, Exception>> connectStream(
  String method,
  String urlStr,
  String body,
) async {
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
    final Uri url = Uri.parse(urlStr);
    final client = http.Client();
    final request = http.Request(method, url)..headers.addAll(getHeader(token));
    request.body = body;
    final response = await client.send(request);
    if (response.statusCode == 200) {
      return Success(response.stream.toStringStream());
    } else {
      return Failure(
        Exception(
          wrapResponseMsg(
            http.Response("", response.statusCode),
            message: 'Error processing tenant',
          ),
        ),
      );
    }
  }
}

Future<Result<void, Exception>> uploadImage(
  PlatformFile image,
  String tenant,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/tenants/$tenant/logo');
    final request = http.MultipartRequest("POST", url);
    request.headers.addAll(getHeader(token));
    request.files.add(
      http.MultipartFile.fromBytes(
        "file",
        image.bytes!,
        filename: image.name,
      ),
    );

    final response = await request.send();
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final String errorMsg = await response.stream.bytesToString();
      return Failure(Exception(errorMsg));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<dynamic, Exception>> backupTenantDB(
  String tenantName,
  String password,
  bool shouldDownload,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/tenants/$tenantName/backup');
    final response = await http.post(
      url,
      body: json.encode(<String, dynamic>{
        'password': password,
        'shouldDownload': shouldDownload,
      }),
      headers: getHeader(token),
    );
    if (response.statusCode == 200) {
      if (shouldDownload) {
        return Success(response.bodyBytes);
      } else {
        return Success(response.body);
      }
    } else {
      final String data = json.decode(response.body);
      return Failure(
        Exception(
          wrapResponseMsg(response, message: "Error backing up tenant $data"),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<String, Exception>> restoreTenantDB(
  PlatformFile backup,
  String tenantName,
  String password,
  bool shouldDrop,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/tenants/$tenantName/restore');
    final request = http.MultipartRequest("POST", url);
    request.fields['password'] = password;
    request.fields['shouldDrop'] = shouldDrop.toString();
    request.headers.addAll(getHeader(token));
    request.files.add(
      http.MultipartFile.fromBytes(
        "file",
        backup.bytes!,
        filename: backup.name,
      ),
    );
    final response = await request.send();
    if (response.statusCode == 200) {
      final String msg = await response.stream.bytesToString();
      return Success(msg);
    } else {
      final String errorMsg = await response.stream.bytesToString();
      return Failure(Exception(errorMsg));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createBackendServer(
  Map<String, dynamic> newBackend,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/servers');
    final response = await http.post(
      url,
      body: json.encode(newBackend),
      headers: getHeader(token),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      return Failure(
        Exception(
          wrapResponseMsg(
            response,
            message: "Error creating backend: ${response.body}",
          ),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> deleteTenant(
  String objName, {
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$apiUrl/api/tenants/$objName');
    final response = await client.delete(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      return Failure(
        Exception(
          wrapResponseMsg(
            response,
            message: "Error deleting tenant: ${response.body}",
          ),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<DockerContainer>, Exception>> fetchTenantDockerInfo(
  String tenantName, {
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$apiUrl/api/tenants/$tenantName');
    final response = await client.get(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      final List<dynamic> data = json.decode(response.body);
      final List<DockerContainer> converted = [];
      for (final item in data) {
        converted.add(DockerContainer.fromMap(item));
      }
      return Success(converted);
    } else {
      return Failure(
        Exception(
          wrapResponseMsg(
            response,
            message: "Error backing up tenant ${response.body}",
          ),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<String, Exception>> fetchContainerLogs(
  String name, {
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$apiUrl/api/containers/$name');
    final response = await client.get(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      final Map<String, dynamic> data = json.decode(response.body);
      return Success(data["logs"].toString());
    } else {
      return Failure(
        Exception(wrapResponseMsg(response, message: "Failed to load logs")),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createNetbox(Nbox netbox) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/tools/netbox');
    final response =
        await http.post(url, body: netbox.toJson(), headers: getHeader(token));
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final String data = json.decode(response.body);
      return Failure(
        Exception(
          wrapResponseMsg(response, message: "Error creating netbox $data"),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createNautobot(Nbox nautobot) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/tools/nautobot');
    final response = await http.post(
      url,
      body: nautobot.toJson(),
      headers: getHeader(token),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final String data = json.decode(response.body);
      return Failure(
        Exception(
          wrapResponseMsg(response, message: "Error creating nautobot $data"),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createOpenDcim(
  String dcimPort,
  adminerPort,
) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/tools/opendcim');
    final response = await http.post(
      url,
      body: json.encode(<String, dynamic>{
        'dcimPort': dcimPort,
        'adminerPort': adminerPort,
      }),
      headers: getHeader(token),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final String data = json.decode(response.body);
      return Failure(
        Exception(
          wrapResponseMsg(response, message: "Error creating netbox $data"),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> deleteTool(String tool) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/tools/$tool');
    final response = await http.delete(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final String data = json.decode(response.body);
      return Failure(
        Exception(
          wrapResponseMsg(
            response,
            message: "Error creating application $data",
          ),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> uploadNetboxDump(PlatformFile file) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/tools/netbox/dump');
    final request = http.MultipartRequest("POST", url);
    request.headers.addAll(getHeader(token));
    request.files.add(
      http.MultipartFile.fromBytes("file", file.bytes!, filename: file.name),
    );
    final response = await request.send();
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final String errorMsg = await response.stream.bytesToString();
      return Failure(
        Exception(
          wrapResponseMsg(
            http.Response(errorMsg, response.statusCode),
            message: errorMsg,
          ),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> importNetboxDump() async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/tools/netbox/import');
    final response = await http.post(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final String data = json.decode(response.body);
      return Failure(
        Exception(
          wrapResponseMsg(
            response,
            message: "Error importing netbox dump: $data",
          ),
        ),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Map<String, dynamic>, Exception>> fetchSchema(String id) async {
  try {
    final Uri url = Uri.parse('$apiUrl/api/schemas/$id');
    final response = await http.get(url, headers: getHeader(token));
    if (response.statusCode == 200) {
      final Map<String, dynamic> data = json.decode(response.body);
      return Success(data);
    } else {
      final Map<String, dynamic> data = json.decode(response.body);
      return Failure(Exception(data["message"].toString()));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}
