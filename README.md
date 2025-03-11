## STC-OTG-SETUP

This document describes deploying STC Labserver and OTG services using a docker-compose yaml file and its required the below files.

  - **otg-compose.yaml**: Main docker compose yaml file
  - **.env**: environment variables defined for docker compose file.
  - **Dockerfile**: Dockerfile to build otg services
  - **entrypoint.sh**: shell script used to start otg services(gnmi and otg)
  - **otg-multi-compose.yaml**: docker compose file to start multi instances of otg services

## Deployment steps
 1. Clone the repository
      `git clone https://github.com/SpirentOrion/stc-otg-setup`
 2. Download and keep the otg binary file(otgservice.VXXX.sh) in stc-otg-setup folder
 3. Download and extract labserver container image locally if required any particular labserver version(labserver-[version number].tar.xz)
      `docker load -i labserver-5.49.2816.tar.xz`
 4. Update the environment variables and otgservice binary file in .env file
 5. Remove if any existing otg docker image using `docker rmi <imageid>`
 6. Run the compose file to deploy otg and labserver services as below.
      `docker-compose -f otg-compose.yaml up -d`
 7. Use **otg-multi-compose.yaml** file to start multi instances of otg service
      `docker-compose -f otg-multi-compose.yaml up --scale otg=2 -d`
    **Note:**
    **--scale otg=<no of otg instances>** option used to create multiple instances of otg/gnmi service and host ports will be assigned dynamically in below range.
    **otg service:** 48153-48200
    **gnmi service:** 49153-49200

 8. Check if docker containers are running
      `docker ps -a`
 9. Login into otg container and check if gnmi and otg services are started
      `docker exec -it otg /bin/bash`
 10. Stop the containers.
      `docker-compose -f otg-compose.yaml down`

## Environment
  - Install docker engine and docker compose on any flavour of Linux VM.
      **Docker version**: 27.3.1
      **Docker-compose version**: 1.29.2

## AION Licensing
  - **Create or update AION user:**<br>
    `curl -X PUT http://${OTGSERVICE_TARGET}/user -H "Content-Type: application/json" -d '{"username":"","server":"","password":"","workspace":"" }'`

    For example:<br>
    `curl -X PUT http://localhost:50051/user -H "Content-Type: application/json" -d '{"username":"demo@spirent.com","server":"http://10.61.37.138/","password":"","workspace":"Spirent"}'`
  - **Get aion user:**<br>
    `curl -X GET http://${OTGSERVICE_TARGET}/user`
  - **Remove aion user:**<br>
    `curl -X POST http://${OTGSERVICE_TARGET}/user/remove`
  - **Test aion user connection:**<br>
     `curl -X POST http://${OTGSERVICE_TARGET}/user/test`

## ToDo
    1.KNE deployment solution.
