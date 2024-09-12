// ignore_for_file: public_member_api_docs, sort_constructors_first
import 'dart:convert';

class Project {
  final String? id;
  String name;
  String dateRange;
  String namespace;
  String authorLastUpdate;
  String lastUpdate;
  bool showAvg;
  bool showSum;
  final bool isPublic;
  List<String> attributes;
  List<String> objects;
  final List<String> permissions;
  bool isImpact;

  Project(
      this.name,
      this.dateRange,
      this.namespace,
      this.authorLastUpdate,
      this.lastUpdate,
      this.showAvg,
      this.showSum,
      this.isPublic,
      this.attributes,
      this.objects,
      this.permissions,
      {this.id,
      this.isImpact = false,});

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      'name': name,
      'dateRange': dateRange,
      'namespace': namespace,
      'authorLastUpdate': authorLastUpdate,
      'lastUpdate': lastUpdate,
      'showAvg': showAvg,
      'showSum': showSum,
      'isPublic': isPublic,
      'attributes': attributes,
      'objects': objects,
      'permissions': permissions,
      'isImpact': isImpact,
    };
  }

  factory Project.fromMap(Map<String, dynamic> map) {
    return Project(
      map['name'].toString(),
      map['dateRange'].toString(),
      map['namespace'].toString(),
      map['authorLastUpdate'].toString(),
      map['lastUpdate'].toString(),
      map['showAvg'] as bool,
      map['showSum'] as bool,
      map['isPublic'] as bool,
      List<String>.from(map['attributes']),
      List<String>.from(map['objects']),
      List<String>.from(map['permissions']),
      id: map['Id'].toString(),
      isImpact: map['isImpact'] is bool ? map['isImpact'] as bool : false,
    );
  }

  String toJson() => json.encode(toMap());

  factory Project.fromJson(String source) =>
      Project.fromMap(json.decode(source) as Map<String, dynamic>);
}
