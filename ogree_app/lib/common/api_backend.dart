import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';
import 'package:http/http.dart' as http;
import 'package:ogree_app/models/domain.dart';
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/models/tenant.dart';
import 'package:ogree_app/models/user.dart';

part 'api_tenant.dart';

String apiUrl = "";
String tenantUrl = "";
const String apiUrlEnvSet = String.fromEnvironment(
  'API_URL',
  defaultValue: 'http://localhost:3001',
);
var token = "";
var tenantToken = "";
getHeader(token) => {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Authorization': 'Bearer $token',
    };

Future<List<String>> loginAPI(String email, String password,
    {String userUrl = ""}) async {
  if (userUrl != "") {
    apiUrl = userUrl;
  } else {
    apiUrl = apiUrlEnvSet;
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
    return [data["email"].toString(), data["isTenant"] ?? ""];
  } else {
    return [""];
  }
}

Future<List<Map<String, List<String>>>> fetchObjectsTree(
    {onlyDomain = false}) async {
  print("API get tree");
  String localUrl = '$apiUrl/api/hierarchy';
  String localToken = token;
  if (onlyDomain) {
    localUrl = '$tenantUrl/api/hierarchy/domains';
    localToken = tenantToken;
  }
  Uri url = Uri.parse(localUrl);
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

Future<String> createBackendServer(Map<String, String> newBackend) async {
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

Future<List<Map<String, String>>> fetchTenantDockerInfo(String tenantName,
    {http.Client? client}) async {
  print("API get Tenant Docker Info");
  client ??= http.Client();
  Uri url = Uri.parse('$apiUrl/api/tenants/${tenantName.toLowerCase()}');
  final response = await client.get(url, headers: getHeader(token));
  print(response.statusCode);
  if (response.statusCode == 200) {
    print(response.body);
    List<dynamic> data = json.decode(response.body);
    print("response.body");
    List<Map<String, String>> converted = [];
    for (var item in data) {
      converted.add(Map<String, String>.from(item));
    }
    print(converted);
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
