import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:mockito/annotations.dart';
import 'package:mockito/mockito.dart';
import 'package:ogree_app/common/api_backend.dart';
import 'package:ogree_app/common/definitions.dart';
import 'package:ogree_app/models/project.dart';

import 'api_test.mocks.dart';

const projectSample =
    '{"data":{"projects":[{"Id":"123xxx","name":"test","dateRange":"21/02/2023","namespace":"Physique","attributes":["vendor","heightUnit","slot","posXY","weigth"],"objects":["site.building.room.rack","site.building.room.rack.device"],"permissions":["user@email.com","admin"],"authorLastUpdate":"Admin","lastUpdate":"21/02/2023","showAvg":true,"showSum":true,"isPublic":false}]}}';

@GenerateMocks([http.Client])
void main() {
  group('fetchProjects', () {
    test('returns a list of Projects if the http call completes successfully',
        () async {
      final mockClient = MockClient();

      // Use Mockito to return a successful response when it calls the
      // provided http.Client.
      when(mockClient.get(Uri.parse('$apiUrl/api/projects?user=user@email.com'),
              headers: getHeader(token),),)
          .thenAnswer((_) async => http.Response(projectSample, 200));

      expect(await fetchProjects("user@email.com", client: mockClient),
          isA<Success<List<Project>, Exception>>(),);
    });

    test('throws an exception if the http call completes with an error',
        () async {
      final mockClient = MockClient();

      // Use Mockito to return an unsuccessful response when it calls the
      // provided http.Client.
      when(mockClient.get(Uri.parse('$apiUrl/api/projects?user=user@email.com'),
              headers: getHeader(token),),)
          .thenAnswer((_) async => http.Response('Not Found', 404));

      expect(await fetchProjects("user@email.com", client: mockClient),
          isA<Failure<List<Project>, Exception>>(),);
    });
  });
}
