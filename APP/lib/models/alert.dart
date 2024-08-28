// ignore_for_file: public_member_api_docs, sort_constructors_first
import 'dart:convert';

class Alert {
  String id;
  String type;
  String title;
  String subtitle;

  Alert(
    this.id,
    this.type,
    this.title,
    this.subtitle,
  );

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      'id': id,
      'type': type,
      'title': title,
      'subtitle': subtitle,
    };
  }

  factory Alert.fromMap(Map<String, dynamic> map) {
    return Alert(
      map['id'].toString(),
      map['type'].toString(),
      map['title'].toString(),
      map['subtitle'].toString(),
    );
  }

  String toJson() => json.encode(toMap());

  factory Alert.fromJson(String source) =>
      Alert.fromMap(json.decode(source) as Map<String, dynamic>);
}
