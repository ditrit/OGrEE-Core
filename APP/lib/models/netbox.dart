import 'dart:convert';

enum Tools { netbox, nautobot, opendcim, cli, unity }

// Nbox applies to both netbox and nautobot
class Nbox {
  String userName;
  String userPassword;
  String port;
  String? version;

  Nbox(this.userName, this.userPassword, this.port);

  Map<String, dynamic> toMap() {
    if (version != null) {
      return <String, dynamic>{
        'username': userName,
        'password': userPassword,
        'port': port,
        'version': version,
      };
    }
    return <String, dynamic>{
      'username': userName,
      'password': userPassword,
      'port': port,
    };
  }

  factory Nbox.fromMap(Map<String, dynamic> map) {
    final nbox = Nbox(
      map['username'].toString(),
      map['password'].toString(),
      map['port'].toString(),
    );
    if (map['version'] != null) {
      nbox.version = map['version'].toString();
    }
    return nbox;
  }

  String toJson() => json.encode(toMap());

  factory Nbox.fromJson(String source) =>
      Nbox.fromMap(json.decode(source) as Map<String, dynamic>);
}
