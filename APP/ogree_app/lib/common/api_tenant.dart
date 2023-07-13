part of 'api_backend.dart';

Future<Result<void, Exception>> loginAPITenant(
    String email, String password, String userUrl) async {
  print("API login to ogree-api $userUrl");
  try {
    Uri url = Uri.parse('$userUrl/api/login');
    final response = await http.post(url,
        body: json
            .encode(<String, String>{'email': email, 'password': password}));
    if (response.statusCode == 200) {
      Map<String, dynamic> data = json.decode(response.body);
      data = (Map<String, dynamic>.from(data["account"]));
      tenantUrl = userUrl;
      tenantToken = data["token"]!;
      return const Success(null);
    } else {
      print(response.statusCode);
      return Failure(Exception(""));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Map<String, dynamic>, Exception>> fetchTenantStats(
    {http.Client? client}) async {
  print("API get Tenant Stats $tenantUrl");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$tenantUrl/api/stats');
    final response = await client.get(url, headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode == 200) {
      print(response.body);
      Map<String, dynamic> data = json.decode(response.body);
      return Success(data);
    } else {
      return Failure(Exception('${response.statusCode}: Failed to load stats'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Map<String, dynamic>, Exception>> fetchTenantApiVersion(
    {http.Client? client}) async {
  print("API get Tenant Version $tenantUrl");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$tenantUrl/api/version');
    final response = await client.get(url, headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode == 200) {
      print(response.body);
      Map<String, dynamic> data = json.decode(response.body);
      data = (Map<String, dynamic>.from(data["data"]));
      return Success(data);
    } else {
      return Failure(
          Exception('${response.statusCode}: Failed to load version'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<User>, Exception>> fetchApiUsers(
    {http.Client? client}) async {
  print("API get users $tenantUrl");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$tenantUrl/api/users');
    final response = await client.get(url, headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode == 200) {
      print(response.body);
      Map<String, dynamic> data = json.decode(response.body);
      print(data["data"]);
      print(data["data"].runtimeType);
      List<User> users = [];
      for (var user in List<Map<String, dynamic>>.from(data["data"])) {
        users.add(User.fromMap(user));
      }
      print(users);
      return Success(users);
    } else {
      return Failure(Exception('${response.statusCode}: Failed to load users'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createUser(User user) async {
  print("API create User");
  try {
    Uri url = Uri.parse('$tenantUrl/api/users');
    final response = await http.post(url,
        body: user.toJson(), headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode == 201) {
      return const Success(null);
    } else {
      var data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> modifyUser(
    String id, Map<String, String> roles) async {
  print("API modify User");
  try {
    Uri url = Uri.parse('$tenantUrl/api/users/$id');
    final response = await http.patch(url,
        body: json.encode(<String, dynamic>{
          'roles': roles,
        }),
        headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      var data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createDomain(Domain domain) async {
  print("API create Domain");
  try {
    Uri url = Uri.parse('$tenantUrl/api/domains');
    final response = await http.post(url,
        body: domain.toJson(), headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode == 201) {
      return const Success(null);
    } else {
      var data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<String, Exception>> createBulkFile(
    Uint8List file, String type) async {
  print("API create bulk $type");
  try {
    Uri url = Uri.parse('$tenantUrl/api/$type/bulk');
    final response =
        await http.post(url, body: file, headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode == 200) {
      var data = json.decode(response.body);
      print(data.toString());
      return Success(data.toString());
    } else {
      var data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> removeObject(String objName, String objType,
    {http.Client? client}) async {
  print("API delete object $objType");
  client ??= http.Client();
  try {
    Uri url = Uri.parse('$tenantUrl/api/$objType/$objName');
    final response = await client.delete(url, headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode >= 200 && response.statusCode < 300) {
      return const Success(null);
    } else {
      var data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Domain, Exception>> fetchDomain(String name) async {
  print("API create Domain");
  try {
    Uri url = Uri.parse('$tenantUrl/api/domains/$name');
    final response = await http.get(url, headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode >= 200 && response.statusCode < 300) {
      print(response.body);
      Map<String, dynamic> data = json.decode(response.body);
      Domain domain = Domain.fromMap(data["data"]);
      return Success(domain);
    } else {
      return Failure(Exception("Unable to load domain"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> updateDomain(
    String currentDomainId, Domain domain) async {
  print("API update Domain");
  try {
    Uri url = Uri.parse('$tenantUrl/api/domains/$currentDomainId');
    final response = await http.put(url,
        body: domain.toJson(), headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      var data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> updateUser(User user) async {
  print("API update Domain");
  try {
    Uri url = Uri.parse('$tenantUrl/api/domains/${user.id}');
    final response = await http.put(url,
        body: user.toJson(), headers: getHeader(tenantToken));
    print(response.statusCode);
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      var data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}
