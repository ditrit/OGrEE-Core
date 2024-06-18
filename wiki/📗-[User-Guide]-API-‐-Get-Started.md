For information on how to build and run the API check the README from the API folder. Once running, the best way to interact with it is through the CLI or APP, but you can also send HTTP request directly to it. The endpoints available can be found here in this [swagger documentation](https://apidoc.ogree.ditrit.io/). 

A great way to directly interact with the API is to use our **Postman** collection, it can be found under `API/resources/postman`. Here are the main steps to use it:
- Import the collection and local environment.
- Set the environment to point to your API and select as the active environment.
- Under the `User` folder of the collection, use the RBAC Login POST to login. This will set the token for the Authentication header of all other requests.
- Now you are ready to send any of the other requests! Follow the `Populate DB` order of requests to create an hierarchy of objects.

**For developers:** The API is implemented following the [controller-service-repository](https://tom-collings.medium.com/controller-service-repository-16e29a4684e5) pattern. Basically, our controllers receive and handle the HTTP requests, calling services. Our services (`models/` folder under API) implement the business logic (validate objects, check and apply dependencies, etc.) and call repository. Repository has all the necessary interaction with MongoDB to update the database.