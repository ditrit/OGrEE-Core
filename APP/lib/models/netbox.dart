import 'dart:convert';

enum Tools { netbox, nautobot, opendcim, cli, unity }

// Nbox applies to both netbox and nautobot
class Nbox {
  String userName;
  String userPassword;
  String port;

  Nbox(this.userName, this.userPassword, this.port);

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      'username': userName,
      'password': userPassword,
      'port': port,
    };
  }

  factory Nbox.fromMap(Map<String, dynamic> map) {
    return Nbox(
      map['username'].toString(),
      map['password'].toString(),
      map['port'].toString(),
    );
  }

  String toJson() => json.encode(toMap());

  factory Nbox.fromJson(String source) =>
      Nbox.fromMap(json.decode(source) as Map<String, dynamic>);
}
