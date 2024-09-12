import 'dart:convert';

class Tag {
  String slug;
  String description;
  String color;
  String image;

  Tag({
    required this.slug,
    required this.description,
    required this.color,
    required this.image,
  });

  Map<String, String> toMap() {
    return <String, String>{
      'description': description,
      'slug': slug,
      'color': color,
      'image': image,
    };
  }

  factory Tag.fromMap(Map<String, dynamic> map) {
    return Tag(
        description: map['description'].toString(),
        slug: map['slug'].toString(),
        color: map['color'].toString(),
        image: map['image'].toString(),);
  }

  String toJson() => json.encode(toMap());

  factory Tag.fromJson(String source) =>
      Tag.fromMap(json.decode(source) as Map<String, dynamic>);
}
