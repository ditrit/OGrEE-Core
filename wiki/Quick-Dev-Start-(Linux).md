On this page you will see how to set up your linux development environment, including local deployment of the API, DB, APP and swagger, compiling the CLI and running 3D. The same steps apply for other operating systems but the commands may differ:

1. Download go: visit https://go.dev/doc/install
2. Install docker: visit https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository
3. Do the local deploy of API, DB, APP and swagger with docker:

    ```bash
    docker compose -p mytenant --profile web --profile doc -f deploy/docker/docker-compose.yml up
    ```

    This step can also be performed using the SuperAdmin backend. For details visit https://github.com/ditrit/OGrEE-Core/tree/main/deploy#readme
4. Compile the CLI:

    a. `cd CLI/`

    b. `make`

    For details visit https://github.com/ditrit/OGrEE-Core/tree/main/CLI#building
5. (optional) Create an alias to avoid typing the password at every CLI launch:

    ```bash
    echo "alias cli='./cli --user admin --password admin'" >> ~/.bashrc
    ```
6. Copy the file `OGrEE-Core/config-example.toml` as `OGrEE-Core/config.toml`
7. Launch the CLI with `cli`
8. Download the latest version of OGrEE-3D from https://github.com/ditrit/OGrEE-3D/releases
9. Unzip the downloaded file
10. Enter the unzipped folder
11. `chmod +x OGrEE-3D.x86_64`
12. Launch the 3D with `./OGrEE-3D.x86_64`

This way, you will be ready to run OGrEE-CLI and OGrEE-3D.

Also, if you are part of the OGrEE team, you will also have access to the demos in Nextcloud. To use them:

1. Install nextcloud-desktop
2. Connect to Ogree's nextcloud: https://nextcloud.ditrit.io/
3. Make a folder sync that includes at least `Ogree/0_modeles3D` and `Ogree/4_customers`
4. Create a `YOURNAME_root.ocli` file that will allow you to run the different demos:

    a. Make a copy of one of the root files found in `Ogree/4_customers`

    b. Modify the line starting with `.var:ROOT={path}`, replacing `{path}` with the location of the Ogree's nextcloud folder on your computer

5. In the CLI, run `.cmds:{path}`, where `{path}` is the path to the root file created in the previous step. This command will run the demo creation, which you can interact with via the CLI. Run the `draw` command to view the demo in 3D