// ignore_for_file: public_member_api_docs, sort_constructors_first
import 'dart:convert';

class Domain {
  String name;
  String color;
  String description;
  String parent;

  Domain(this.name, this.color, this.description, this.parent);

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      'name': name,
      'attributes': <String, String>{'color': color},
      'category': 'domain',
      'description': description,
      'parentId': parent,
    };
  }

  factory Domain.fromMap(Map<String, dynamic> map) {
    return Domain(
      map['name'].toString(),
      map['attributes']['color'].toString(),
      map['description'].toString(),
      map['parentId'] == null ? "" : map['parentId'].toString(),
    );
  }

  String toJson() => json.encode(toMap());

  factory Domain.fromJson(String source) =>
      Domain.fromMap(json.decode(source) as Map<String, dynamic>);
}
