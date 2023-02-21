import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:ogree_app/models/project.dart';

const URL = "http://localhost:3001/api";
const token = "INSERT TOKEN HERE";
const header = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
  'Authorization': 'Bearer $token',
};

Future<List<Map<String, List<String>>>> fetchObjectsTree() async {
  print("API get tree");
  Uri url = Uri.parse('$URL/hierarchy');
  final response = await http.get(url, headers: header);
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
  Uri url = Uri.parse('$URL/hierarchy/attributes');
  final response = await http.get(url, headers: header);
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

Future<List<Project>> fetchProjects() async {
  print("API get Projects");
  Uri url = Uri.parse('$URL/projects?userid=63a33a07e7e6939da7378204');
  final response = await http.get(url, headers: header);
  print(response.statusCode);
  if (response.statusCode == 200) {
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
  Uri url = Uri.parse('$URL/projects/$id');
  final response = await http.delete(url, headers: header);
  if (response.statusCode == 200) {
    return "";
  } else {
    final Map<String, dynamic> data = json.decode(response.body);
    return data["message"].toString();
  }
}

Future<String> modifyProject(Project project) async {
  print("API modify Projects");
  Uri url = Uri.parse('$URL/projects/${project.id}');
  final response = await http.put(url, body: project.toJson(), headers: header);
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
  Uri url = Uri.parse('$URL/projects');
  final response =
      await http.post(url, body: project.toJson(), headers: header);
  print(response);
  if (response.statusCode == 200) {
    return "";
  } else {
    final Map<String, dynamic> data = json.decode(response.body);
    return data["message"].toString();
  }
}

Future<bool> loginAPI(String email, String password) async {
  print("API login");
  Uri url = Uri.parse('http://127.0.0.1:3001/api/login');
  final response = await http.post(url,
      body:
          json.encode(<String, String>{'email': email, 'password': password}));
  if (response.statusCode == 200) {
    return true;
  } else {
    return false;
  }
}
