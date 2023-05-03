// ignore_for_file: public_member_api_docs, sort_constructors_first
import 'dart:convert';

class Tenant {
  String name;
  String customerPassword;
  String apiUrl;
  String webUrl;
  String apiPort;
  String webPort;
  bool hasWeb;
  bool hasCli;
  bool hasDoc;
  String docUrl;
  String docPort;

  Tenant(
      this.name,
      this.customerPassword,
      this.apiUrl,
      this.webUrl,
      this.apiPort,
      this.webPort,
      this.hasWeb,
      this.hasCli,
      this.hasDoc,
      this.docUrl,
      this.docPort);

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      'name': name,
      'customerPassword': customerPassword,
      'apiUrl': apiUrl,
      'webUrl': webUrl,
      'apiPort': apiPort,
      'webPort': webPort,
      'hasWeb': hasWeb,
      'hasCli': hasCli,
      'hasDoc': hasDoc,
      'docUrl': docUrl,
      'docPort': docPort,
    };
  }

  factory Tenant.fromMap(Map<String, dynamic> map) {
    return Tenant(
        map['name'].toString(),
        map['customerPassword'].toString(),
        map['apiUrl'].toString(),
        map['webUrl'].toString(),
        map['apiPort'].toString(),
        map['webPort'].toString(),
        map['hasWeb'],
        map['hasCli'],
        map['hasDoc'],
        map['docUrl'].toString(),
        map['docPort'].toString());
  }

  String toJson() => json.encode(toMap());

  factory Tenant.fromJson(String source) =>
      Tenant.fromMap(json.decode(source) as Map<String, dynamic>);
}
