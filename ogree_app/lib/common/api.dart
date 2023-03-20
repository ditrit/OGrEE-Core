import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:ogree_app/models/project.dart';
import 'package:ogree_app/models/tenant.dart';

const String apiUrl = String.fromEnvironment(
  'API_URL',
  // defaultValue: 'http://localhost:8081',
  defaultValue: 'https://b.api.ogree.ditrit.io',
);
var token = "";
getHeader(token) => {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'Authorization': 'Bearer $token',
    };

Future<List<Map<String, List<String>>>> fetchObjectsTree() async {
  print("API get tree");
  Uri url = Uri.parse('$apiUrl/api/hierarchy');
  final response = await http.get(url, headers: getHeader(token));
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
    for (var item in converted["categories"]!.keys) {
      categories[item.toString()] =
          List<String>.from(converted["categories"]![item]);
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

Future<Map<String, dynamic>> fetchTenantStats(String tenantUrl,
    {http.Client? client}) async {
  print("API get Tenant Stats $tenantUrl");
  client ??= http.Client();
  Uri url = Uri.parse('$tenantUrl/api/stats');
  final response = await client.get(url);
  print(response.statusCode);
  if (response.statusCode == 200) {
    print(response.body);
    Map<String, dynamic> data = json.decode(response.body);
    return data;
  } else {
    // If the server did not return a 200 OK response,
    // then throw an exception.
    throw Exception('${response.statusCode}: Failed to load objects');
  }
}

Future<Map<String, dynamic>> fetchTenantApiVersion(String tenantUrl,
    {http.Client? client}) async {
  print("API get Tenant Version $tenantUrl");
  client ??= http.Client();
  Uri url = Uri.parse('$tenantUrl/api/version');
  final response = await client.get(url);
  print(response.statusCode);
  if (response.statusCode == 200) {
    print(response.body);
    Map<String, dynamic> data = json.decode(response.body);
    data = (Map<String, dynamic>.from(data["data"]));
    return data;
  } else {
    // If the server did not return a 200 OK response,
    // then throw an exception.
    throw Exception('${response.statusCode}: Failed to load objects');
  }
}

Future<String> loginAPI(String email, String password) async {
  print("API login $apiUrl");
  Uri url = Uri.parse('$apiUrl/api/login');
  final response = await http.post(url,
      body:
          json.encode(<String, String>{'email': email, 'password': password}));
  if (response.statusCode == 200) {
    Map<String, dynamic> data = json.decode(response.body);
    data = (Map<String, dynamic>.from(data["account"]));
    token = data["token"]!;
    return data["Email"].toString();
  } else {
    return "";
  }
}
