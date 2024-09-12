part of 'api_backend.dart';

// This API is used by SuperAdmin mode to connect to multiple tenant APIs
// still being connected to the SuperAdmin backend
Future<Result<void, Exception>> loginAPITenant(
  String email,
  String password,
  String userUrl,
) async {
  try {
    final Uri url = Uri.parse('$userUrl/api/login');
    final response = await http.post(
      url,
      body: json.encode(<String, String>{'email': email, 'password': password}),
    );
    if (response.statusCode == 200) {
      Map<String, dynamic> data = json.decode(response.body);
      data = Map<String, dynamic>.from(data["account"]);
      tenantUrl = userUrl;
      tenantToken = data["token"]!;
      return const Success(null);
    } else {
      return Failure(Exception());
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Map<String, dynamic>, Exception>> fetchTenantStats({
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$tenantUrl/api/stats');
    final response = await client.get(url, headers: getHeader(tenantToken));
    if (response.statusCode == 200) {
      final Map<String, dynamic> data = json.decode(response.body);
      return Success(data);
    } else {
      return Failure(Exception('${response.statusCode}: Failed to load stats'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Map<String, dynamic>, Exception>> fetchTenantApiVersion({
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$tenantUrl/api/version');
    final response = await client.get(url, headers: getHeader(tenantToken));
    if (response.statusCode == 200) {
      Map<String, dynamic> data = json.decode(response.body);
      data = Map<String, dynamic>.from(data["data"]);
      return Success(data);
    } else {
      return Failure(
        Exception('${response.statusCode}: Failed to load version'),
      );
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<User>, Exception>> fetchApiUsers({
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$tenantUrl/api/users');
    final response = await client.get(url, headers: getHeader(tenantToken));
    if (response.statusCode == 200) {
      final Map<String, dynamic> data = json.decode(response.body);
      final List<User> users = [];
      for (final user in List<Map<String, dynamic>>.from(data["data"])) {
        users.add(User.fromMap(user));
      }
      return Success(users);
    } else {
      return Failure(Exception('${response.statusCode}: Failed to load users'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createUser(User user) async {
  try {
    final Uri url = Uri.parse('$tenantUrl/api/users');
    final response = await http.post(
      url,
      body: user.toJson(),
      headers: getHeader(tenantToken),
    );
    if (response.statusCode == 201) {
      return const Success(null);
    } else {
      final data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> modifyUser(
  String id,
  Map<String, String> roles,
) async {
  try {
    final Uri url = Uri.parse('$tenantUrl/api/users/$id');
    final response = await http.patch(
      url,
      body: json.encode(<String, dynamic>{
        'roles': roles,
      }),
      headers: getHeader(tenantToken),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createDomain(Domain domain) async {
  try {
    final Uri url = Uri.parse('$tenantUrl/api/domains');
    final response = await http.post(
      url,
      body: domain.toJson(),
      headers: getHeader(tenantToken),
    );
    if (response.statusCode == 201) {
      return const Success(null);
    } else {
      final data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<String, Exception>> createBulkFile(
  Uint8List file,
  String type,
) async {
  try {
    final Uri url = Uri.parse('$tenantUrl/api/$type/bulk');
    final response =
        await http.post(url, body: file, headers: getHeader(tenantToken));
    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      return Success(data.toString());
    } else {
      final data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> removeObject(
  String objName,
  String objType, {
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$tenantUrl/api/$objType/$objName');
    final response = await client.delete(url, headers: getHeader(tenantToken));
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

Future<Result<Domain, Exception>> fetchDomain(String name) async {
  try {
    final Uri url = Uri.parse('$tenantUrl/api/domains/$name');
    final response = await http.get(url, headers: getHeader(tenantToken));
    if (response.statusCode >= 200 && response.statusCode < 300) {
      final Map<String, dynamic> data = json.decode(response.body);
      final Domain domain = Domain.fromMap(data["data"]);
      return Success(domain);
    } else {
      return Failure(Exception("Unable to load domain"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> updateDomain(
  String currentDomainId,
  Domain domain,
) async {
  try {
    final Uri url = Uri.parse('$tenantUrl/api/domains/$currentDomainId');
    final response = await http.put(
      url,
      body: domain.toJson(),
      headers: getHeader(tenantToken),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> updateUser(User user) async {
  try {
    final Uri url = Uri.parse('$tenantUrl/api/domains/${user.id}');
    final response = await http.put(
      url,
      body: user.toJson(),
      headers: getHeader(tenantToken),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<List<Tag>, Exception>> fetchTags({http.Client? client}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$tenantUrl/api/tags');
    final response = await client.get(url, headers: getHeader(tenantToken));
    if (response.statusCode == 200) {
      final Map<String, dynamic> data = json.decode(response.body);
      final List<Tag> tags = [];
      for (final tag
          in List<Map<String, dynamic>>.from(data["data"]["objects"])) {
        tags.add(Tag.fromMap(tag));
      }
      return Success(tags);
    } else {
      return Failure(Exception('${response.statusCode}: Failed to load tags'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<Tag, Exception>> fetchTag(
  String tagId, {
  http.Client? client,
}) async {
  client ??= http.Client();
  try {
    final Uri url = Uri.parse('$tenantUrl/api/tags/$tagId');
    final response = await client.get(url, headers: getHeader(tenantToken));
    if (response.statusCode == 200) {
      final Map<String, dynamic> data = json.decode(response.body);
      final Tag tag = Tag.fromMap(data["data"]);
      return Success(tag);
    } else {
      return Failure(Exception('${response.statusCode}: Failed to load users'));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> createTag(Tag tag) async {
  try {
    final Uri url = Uri.parse('$tenantUrl/api/tags');
    final response = await http.post(
      url,
      body: tag.toJson(),
      headers: getHeader(tenantToken),
    );
    if (response.statusCode == 201) {
      return const Success(null);
    } else {
      final data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}

Future<Result<void, Exception>> updateTag(
  String currentId,
  Map<String, String> tagMap,
) async {
  try {
    final Uri url = Uri.parse('$tenantUrl/api/tags/$currentId');
    final response = await http.patch(
      url,
      body: json.encode(tagMap),
      headers: getHeader(tenantToken),
    );
    if (response.statusCode == 200) {
      return const Success(null);
    } else {
      final data = json.decode(response.body);
      return Failure(Exception("Error: ${data["message"]}"));
    }
  } on Exception catch (e) {
    return Failure(e);
  }
}
