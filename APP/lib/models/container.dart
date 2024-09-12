import 'dart:convert';

class DockerContainer {
  String name;
  String lastStarted;
  String status;
  String image;
  String size;
  String ports;

  DockerContainer(this.name, this.lastStarted, this.status, this.image,
      this.size, this.ports,);

  Map<String, dynamic> toMap() {
    return <String, dynamic>{
      'Names': name,
      'RunningFor': lastStarted,
      'State': status,
      'Image': image,
      'Size': size,
      'Ports': ports,
    };
  }

  factory DockerContainer.fromMap(Map<String, dynamic> map) {
    return DockerContainer(
      map['Names'].toString(),
      map['RunningFor'].toString(),
      map['State'].toString(),
      map['Image'].toString(),
      map['Size'].toString(),
      map['Ports'].toString(),
    );
  }

  String toJson() => json.encode(toMap());

  factory DockerContainer.fromJson(String source) =>
      DockerContainer.fromMap(json.decode(source) as Map<String, dynamic>);
}
