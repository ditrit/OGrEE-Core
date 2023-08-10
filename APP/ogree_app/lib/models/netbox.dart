import 'dart:convert';

class Netbox {
  String userName;
  String userPassword;
  String port;

  Netbox(this.userName, this.userPassword, this.port);

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      'username': userName,
      'password': userPassword,
      'port': port,
    };
  }

  factory Netbox.fromMap(Map<String, dynamic> map) {
    return Netbox(
      map['username'].toString(),
      map['password'].toString(),
      map['port'].toString(),
    );
  }

  String toJson() => json.encode(toMap());

  factory Netbox.fromJson(String source) =>
      Netbox.fromMap(json.decode(source) as Map<String, dynamic>);
}
