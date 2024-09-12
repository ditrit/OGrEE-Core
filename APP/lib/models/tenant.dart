// ignore_for_file: public_member_api_docs, sort_constructors_first
import 'dart:convert';

enum TenantStatus { running, partialRun, notRunning, unavailable }

class Tenant {
  String name;
  String customerPassword;
  String apiUrl;
  String webUrl;
  String apiPort;
  String webPort;
  bool hasWeb;
  bool hasDoc;
  String docUrl;
  String docPort;
  String imageTag;
  TenantStatus? status;

  Tenant(
      this.name,
      this.customerPassword,
      this.apiUrl,
      this.webUrl,
      this.apiPort,
      this.webPort,
      this.hasWeb,
      this.hasDoc,
      this.docUrl,
      this.docPort,
      this.imageTag,
      {this.status,});

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      'name': name,
      'customerPassword': customerPassword,
      'apiUrl': apiUrl,
      'webUrl': webUrl,
      'apiPort': apiPort,
      'webPort': webPort,
      'hasWeb': hasWeb,
      'hasDoc': hasDoc,
      'docUrl': docUrl,
      'docPort': docPort,
      'imageTag': imageTag,
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
        map['hasDoc'],
        map['docUrl'].toString(),
        map['docPort'].toString(),
        map['imageTag'].toString(),);
  }

  String toJson() => json.encode(toMap());

  factory Tenant.fromJson(String source) =>
      Tenant.fromMap(json.decode(source) as Map<String, dynamic>);
}
