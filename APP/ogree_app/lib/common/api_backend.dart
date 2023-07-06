import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';
import 'package:file_picker/file_picker.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:http/http.dart' as http;
import 'package:ogree_app/models/container.dart';
import 'package:ogree_app/models/domain.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/models/user.dart';

part 'api_tenant.dart';

String apiUrl = "";
String tenantUrl = "";
String tenantName = "";
bool isTenantAdmin = false;
var token = "";
var tenantToken = "";
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
  String urlAppend = "?startDate=${reformatDate(ranges[0])}";
  if (ranges.length > 1) {
    urlAppend = "$urlAppend&endDate=${reformatDate(ranges[1])}";
  }
  return urlAppend;
}

Future<List<String>> loginAPI(String email, String password,
    {String userUrl = ""}) async {
  tenantUrl = "";
  isTenantAdmin = false;
  token = "";
  tenantToken = "";
  if (userUrl != "") {
    apiUrl = userUrl;
  } else {
    apiUrl = dotenv.get('API_URL', fallback: 'http://localhost:3001');
  }
  print("API login ogree $apiUrl");
  Uri url = Uri.parse('$apiUrl/api/login');
  final response = await http.post(url,
      body:
          json.encode(<String, String>{'email': email, 'password': password}));
  if (response.statusCode == 200) {
    Map<String, dynamic> data = json.decode(response.body);
    data = (Map<String, dynamic>.from(data["account"]));
    token = data["token"]!;
    if (data["isTenant"] == null && data["roles"]["*"] == "manager") {
      // Not tenant mode, but tenant admin
      isTenantAdmin = true;
      tenantUrl = apiUrl;
      tenantToken = token;
    }
    return [data["email"].toString(), data["isTenant"] ?? ""];
  } else {
    return [""];
  }
}

Future<bool> fetchApiTenantName({http.Client? client}) async {
  print("API get TenantName");
  client ??= http.Client();
  Uri url = Uri.parse('$apiUrl/api/version');
  final response = await client.get(url, headers: getHeader(token));
  print(response.statusCode);
  if (response.statusCode == 200) {
    Map<String, dynamic> data = json.decode(response.body);
    data = (Map<String, dynamic>.from(data["data"]));
    tenantName = data["Customer"];
    print(tenantName);
    return true;
  }
  return false;
}

Future<String> changeUserPassword(String currentPassword, newPassword) async {
  print("API change password");
  Uri url = Uri.parse('$apiUrl/api/users/password/change');
  final response = await http.post(url,
      body: json.encode(<String, dynamic>{
        'currentPassword': currentPassword,
        'newPassword': newPassword
      }),
      headers: getHeader(token));
  print(response.statusCode);
  if (response.statusCode == 200) {
    Map<String, dynamic> data = json.decode(response.body);
    token = data["token"]!;
    return "";
  } else {
    Map<String, dynamic> data = json.decode(response.body);
    return "Error: ${data["message"]}";
  }
}

Future<String> userForgotPassword(String email, {String userUrl = ""}) async {
  print("API forgot password");
  if (userUrl != "") {
    apiUrl = userUrl;
  } else {
    apiUrl = dotenv.get('API_URL', fallback: 'http://localhost:3001');
  }
  Uri url = Uri.parse('$apiUrl/api/users/password/forgot');
  final response = await http.post(
    url,
    body: json.encode(<String, dynamic>{'email': email}),
  );
  print(response.statusCode);
  if (response.statusCode == 200) {
    Map<String, dynamic> data = json.decode(response.body);
    return "";
  } else {
    Map<String, dynamic> data = json.decode(response.body);
    return "Error: ${data["message"]}";
  }
}

Future<String> userResetPassword(String password, String resetToken,
    {String userUrl = ""}) async {
  print("API reset password");
  if (userUrl != "") {
    apiUrl = userUrl;
  } else {
    apiUrl = dotenv.get('API_URL', fallback: 'http://localhost:3001');
  }
  Uri url = Uri.parse('$apiUrl/api/users/password/reset');
  final response = await http.post(
    url,
    body: json.encode(<String, dynamic>{'newPassword': password}),
    headers: getHeader(resetToken),
  );
  if (response.statusCode == 200) {
    Map<String, dynamic> data = json.decode(response.body);
    print(data);
    return "";
  } else {
    Map<String, dynamic> data = json.decode(response.body);
    return "Error: ${data["message"]}";
  }
}

Future<List<Map<String, List<String>>>> fetchObjectsTree(
    {String dateRange = "",
    bool onlyDomain = false,
    bool isTenantMode = false}) async {
  print("API get tree: onlydomain=$onlyDomain");
  String localUrl = '/api/hierarchy';
  String localToken = token;
  if (isTenantMode) {
    localUrl = tenantUrl + localUrl;
    localToken = tenantToken;
  } else {
    localUrl = apiUrl + localUrl;
  }
  if (onlyDomain) {
    localUrl = '$localUrl/domains';
  }
  if (dateRange != "") {
    localUrl = localUrl + urlDateAppend(dateRange);
  }
  Uri url = Uri.parse(localUrl);
  try {
    final response = await http.get(url, headers: getHeader(localToken));
    print(response.statusCode);
    if (response.statusCode == 200) {
      // If the server did return a 200 OK response,
      // then parse the JSON and convert to the right map format.
      Map<String, dynamic> data = json.decode(response.body);
      data = (Map<String, dynamic>.from(data["data"]));
      Map<String, Map<String, dynamic>> converted = {};
      Map<String, List<String>> tree = {};
      Map<String, List<String>> categories = {};
      for (var item in data.keys) {
        converted[item.toString()] = Map<String, dynamic>.from(data[item]);
      }
      for (var item in converted["tree"]!.keys) {
        tree[item.toString()] = List<String>.from(converted["tree"]![item]);
      }
      if (!onlyDomain) {
        for (var item in converted["categories"]!.keys) {
          categories[item.toString()] =
              List<String>.from(converted["categories"]![item]);
        }
      }
      return [tree, categories];
    } else {
      // If the server did not return a 200 OK response,
      // then throw an exception.
      throw Exception('${response.statusCode}: Failed to load objects');
    }
  } catch (e) {
    print(e);
    throw Exception('Failed to load objects');
  }
}

Future<Map<String, Map<String, String>>> fetchAttributes() async {
  print("API get Attrs");
  Uri url = Uri.parse('$apiUrl/api/hierarchy/attributes');
  final response = await http.get(url, headers: getHeader(token));
  print(response.statusCode);
  if (response.statusCode == 200) {
    // If the server did return a 200 OK response,
    // then parse the JSON and convert to the right map format.
    Map<String, dynamic> data = json.decode(response.body);
    data = (Map<String, dynamic>.from(data["data"]));
    Map<String, Map<String, String>> converted = {};
    for (var item in data.keys) {
      converted[item.toString()] = Map<String, String>.from(data[item]);
    }
    return converted;
  } else {
    // If the server did not return a 200 OK response,
    // then throw an exception.
    throw Exception('${response.statusCode}: Failed to load objects');
  }
}

Future<List<Project>> fetchProjects(String userEmail,
    {http.Client? client}) async {
  print("API get Projects");
  client ??= http.Client();
  Uri url = Uri.parse('$apiUrl/api/projects?user=$userEmail');
  final response = await client.get(url, headers: getHeader(token));
  print(response.statusCode);
  if (response.statusCode == 200) {
    print(response);
    print(response.body);
    // If the server did return a 200 OK response,
    // then parse the JSON and convert to the right format.
    Map<String, dynamic> data = json.decode(response.body);
    data = (Map<String, dynamic>.from(data["data"]));
    List<Project> projects = [];
    for (var project in data["projects"]) {
      projects.add(Project.fromMap(project));
    }
    return projects;
  } else {
    // If the server did not return a 200 OK response,
    // then throw an exception.
    throw Exception('${response.statusCode}: Failed to load objects');
  }
}

Future<String> deleteProject(String id) async {
  print("API delete Projects");
  Uri url = Uri.parse('$apiUrl/api/projects/$id');
  final response = await http.delete(url, headers: getHeader(token));
  if (response.statusCode == 200) {
    return "";
  } else {
    final Map<String, dynamic> data = json.decode(response.body);
    return data["message"].toString();
  }
}

Future<String> modifyProject(Project project) async {
  print("API modify Projects");
  Uri url = Uri.parse('$apiUrl/api/projects/${project.id}');
  final response =
      await http.put(url, body: project.toJson(), headers: getHeader(token));
  print(response);
  if (response.statusCode == 200) {
    return "";
  } else {
    final Map<String, dynamic> data = json.decode(response.body);
    return data["message"].toString();
  }
}

Future<String> createProject(Project project) async {
  print("API create Projects");
  Uri url = Uri.parse('$apiUrl/api/projects');
  final response =
      await http.post(url, body: project.toJson(), headers: getHeader(token));
  print(response);
  if (response.statusCode == 200) {
    return "";
  } else {
    final Map<String, dynamic> data = json.decode(response.body);
    return data["message"].toString();
  }
}

Future<List<Tenant>> fetchTenants({http.Client? client}) async {
  print("API get Tenants");
  client ??= http.Client();
  Uri url = Uri.parse('$apiUrl/api/tenants');
  final response = await client.get(url, headers: getHeader(token));
  print(response.statusCode);
  if (response.statusCode == 200) {
    print(response);
    print(response.body);
    // If the server did return a 200 OK response,
    // then parse the JSON and convert to the right format.
    Map<String, dynamic> data = json.decode(response.body);
    List<Tenant> tenants = [];
    for (var project in data["tenants"]) {
      tenants.add(Tenant.fromMap(project));
    }
    return tenants;
  } else {
    // If the server did not return a 200 OK response,
    // then throw an exception.
    throw Exception('${response.statusCode}: Failed to load objects');
  }
}

Future<String> createTenant(Tenant tenant) async {
  print("API create Tenants");
  Uri url = Uri.parse('$apiUrl/api/tenants');
  final response =
      await http.post(url, body: tenant.toJson(), headers: getHeader(token));
  print(response);
  if (response.statusCode == 200) {
    return "";
  } else {
    String data = json.decode(response.body);
    return "Error creating tenant $data";
  }
}

Future<String> updateTenant(Tenant tenant) async {
  print("API update Tenants");
  Uri url = Uri.parse('$apiUrl/api/tenants/${tenant.name}');
  final response =
      await http.put(url, body: tenant.toJson(), headers: getHeader(token));
  print(response);
  if (response.statusCode == 200) {
    return "";
  } else {
    String data = json.decode(response.body);
    return "Error creating tenant $data";
  }
}

Future<String> uploadImage(PlatformFile image, String tenant) async {
  print("API upload Tenant logo");
  Uri url = Uri.parse('$apiUrl/api/tenants/$tenant/logo');
  var request = http.MultipartRequest("POST", url);
  request.headers.addAll(getHeader(token));
  request.files.add(
      http.MultipartFile.fromBytes("file", image.bytes!, filename: image.name));

  var response = await request.send();
  print(response.statusCode);
  var body = await response.stream.bytesToString();
  return body;
}

Future<String> createBackendServer(Map<String, dynamic> newBackend) async {
  print("API create Back Server");
  Uri url = Uri.parse('$apiUrl/api/servers');
  final response = await http.post(url,
      body: json.encode(newBackend), headers: getHeader(token));
  print(response);
  if (response.statusCode == 200) {
    return "";
  } else {
    return "Error creating backend ${response.body}";
  }
}

Future<String> deleteTenant(String objName, {http.Client? client}) async {
  print("API delete Tenant");
  client ??= http.Client();
  Uri url = Uri.parse('$apiUrl/api/tenants/$objName');
  final response = await client.delete(url, headers: getHeader(token));
  print(response.statusCode);
  if (response.statusCode == 200) {
    return "";
  } else {
    return response.body;
  }
}

Future<List<DockerContainer>> fetchTenantDockerInfo(String tenantName,
    {http.Client? client}) async {
  print("API get Tenant Docker Info");
  client ??= http.Client();
  Uri url = Uri.parse('$apiUrl/api/tenants/${tenantName.toLowerCase()}');
  final response = await client.get(url, headers: getHeader(token));
  print(response.statusCode);
  if (response.statusCode == 200) {
    List<dynamic> data = json.decode(response.body);
    List<DockerContainer> converted = [];
    for (var item in data) {
      converted.add(DockerContainer.fromMap(item));
    }
    return converted;
  } else {
    // If the server did not return a 200 OK response,
    // then throw an exception.
    print('${response.statusCode}: ${response.body}');
    return [];
  }
}

Future<String> fetchContainerLogs(String name, {http.Client? client}) async {
  print("API get Container Logs $name");
  client ??= http.Client();
  Uri url = Uri.parse('$apiUrl/api/containers/$name');
  final response = await client.get(url, headers: getHeader(token));
  print(response.statusCode);
  if (response.statusCode == 200) {
    // print(response.body);
    Map<String, dynamic> data = json.decode(response.body);
    return data["logs"].toString();
  } else {
    // If the server did not return a 200 OK response,
    // then throw an exception.
    throw Exception('${response.statusCode}: Failed to load objects');
  }
}
