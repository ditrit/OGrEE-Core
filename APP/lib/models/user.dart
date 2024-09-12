import 'dart:convert';

const allDomainsTag = "*";
const allDomainsConvert = "All domains";

class User {
  String? id;
  String name;
  String email;
  String password;
  Map<String, String> roles;

  User(
      {required this.name,
      required this.email,
      required this.password,
      required this.roles,
      this.id,});

  Map<String, dynamic> toMap() {
    if (roles[allDomainsConvert] != null) {
      roles[allDomainsTag] = roles[allDomainsConvert]!;
      roles.remove(allDomainsConvert);
    }
    return <String, dynamic>{
      'name': name,
      'email': email,
      'password': password,
      'roles': roles,
    };
  }

  factory User.fromMap(Map<String, dynamic> map) {
    final roles = Map<String, String>.from(map['roles']);
    if (roles[allDomainsTag] != null) {
      roles[allDomainsConvert] = roles[allDomainsTag]!;
      roles.remove(allDomainsTag);
    }
    return User(
        name: map['name'].toString(),
        id: map['_id'].toString(),
        email: map['email'].toString(),
        password: map['password'].toString(),
        roles: roles,);
  }

  String toJson() => json.encode(toMap());

  factory User.fromJson(String source) =>
      User.fromMap(json.decode(source) as Map<String, dynamic>);

  @override
  String toString() {
    return 'User(id: $id, email: $email, password: $password, roles: $roles)';
  }
}
