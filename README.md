# Spirent TestCenter OTG API Support Overview

Spirent TestCenter (STC) is a powerful, high-performance testing platform designed to validate and optimize network infrastructure, protocols, and applications. To enhance automation and integration capabilities, STC supports the OTG (Open Traffic Generator) and gNMI (gRPC Network Management Interface) APIs, enabling users to leverage open-source, standardized test scripts in flexible and scalable open-source test environments.

## Key Benefits of Spirent TestCenter’s OTG API Support

With Spirent TestCenter’s OTG API support, users can:

- Deploy Spirent TestCenter with ease in Ondatra, snappi, and gosnappi test environments
- Control any version of Spirent TestCenter, including appliances, chassis-based hardware, and virtual solutions
- Automate and streamline network traffic and protocols testing with open-source tools
- Leverage a vendor-agnostic approach for multi-platform interoperability
- Easily integrate Spirent TestCenter into existing DevOps and CI/CD environments

To help users set up and configure STC with OTG API, this repository includes comprehensive setup instructions, sample scripts, and best practices for seamless implementation.

## Spirent OTG Service Deployment Guide

The Spirent OTG Service is available for use on Linux PC/workstations and depends on the installation of the Spirent ReST API application. You can pre-install either the Spirent TestCenter application or Spirent Labserver as the ReST API application.

### Recommended System Requirements

- **CPU:** 2 core
- **RAM:** 2 GB
- **Disk:** 30 GB (SSD or better)
- **STC Version:** 5.52 or later
- The OTG Service can be installed on the same machine as the ReST API application.If installing on the same host, refer to the system resource requirements for STC or STC LabServer.
- If using Spirent AION Licensing, refer to the [Spirent TestCenter with Spirent AION Licensing Quick Start Guide (DOC12187)] for further details.

## Deploying OTG Service Directly

### Key Features

- Suitable for standalone installations on a single machine.
- Requires manual running of the installation script and configuring services.
- Best for cases where Docker is not used or when installing the OTG service alongside a ReST API application.

### Installation Steps

#### Step 1: Obtain the Installation File

Download the executable file:

```sh
File Name: otgservice.V[x.xx].sh
```

#### Step 2: Install the OTG Service

Run the installer to extract the files into a new folder with the same name as the package:

```sh
./otgservice.V[x.xx].sh
```

#### Step 3: Start OTG and gNMI Services

Navigate to the extracted folder and start the services using the default `otg.conf` file:

```sh
cd otgservice.V[x.xx]
./otgctl --start
```
#### Step 4: Configure the STC ReST Server IP

Set the ReST server IP address:

```sh
./otgctl --restserver 1.2.3.4:80
```

#### Step 5: Access Logs

Logs are automatically saved in:

```sh
./otgservice.V[x.xx]/Logs/
```

To view all available commands(Optional), run:

```sh
./otgctl --help
```

## Deploying OTG Service as a Docker Container

### Key Features

- Automates deployment using Docker containers for both OTG and Labserver services.
- Supports scalability, allowing multiple OTG instances to be created dynamically.
- Ideal for test labs or cloud environments where services need to be managed and deployed efficiently.

### Prerequisites
- Install Docker engine and Docker Compose on any flavor of Linux VM.
  - **Docker Version:** 27.3.1
  - **Docker Compose Version:** 1.29.2
- **System Resource Requirements:** Refer to the STC Labserver resource requirements.

### Installation Steps

#### Step 1: Clone the Repository

Download the required setup files by cloning the repository:

```sh
git clone https://github.com/SpirentOrion/stc-otg-setup
cd stc-otg-setup
```

Required files:

- `otg-compose.yaml`: Main Docker Compose YAML file
- `.env`: Environment variables defined for Docker Compose file
- `Dockerfile`: Dockerfile to build OTG services
- `entrypoint.sh`: Shell script used to start OTG services (GNMI and OTG)
- `otg-multi-compose.yaml`: Docker Compose file to start multiple instances of OTG services
- `otgservice.V[x.xx]`: Spirent OTG Service application

#### Step 2: Load the Labserver Docker Image

If a specific Labserver version is needed, download and load it into Docker:

```sh
docker load -i labserver-5.52.xxxx.tar.xz
```

#### Step 3: Update Configuration Files

Modify the `.env` file to set the required environment variables and specify the OTG service binary file.

#### Step 4: Deploy OTG & Labserver Services

Run the Docker Compose file to start the services:

```sh
docker-compose -f otg-compose.yaml up -d
```

#### Step 5: Deploy Multiple OTG Instances (Optional)

To run multiple instances of the OTG service, use the following command:

```sh
docker-compose -f otg-multi-compose.yaml up --scale otg=2 -d
```

- The `--scale` option allows you to create multiple OTG/gNMI service instances.
- **Default Dynamic Port Ranges:**
  - **OTG Service:** 48153–48200
  - **gNMI Service:** 49153–49200
  
Use the following command to create multiple instances of the OTG service, including STCv ports and the LabServer. 
This setup is used to validate multiple OTG instances in parallel by running GoSNAPPI and Ondatra test cases.

```sh
docker-compose -f otg-stcv-labserver-compose.yaml up --scale otg=2 --scale stcv=4 -d
```
- The above command will create 2 otg services and 4 stcv ports and 1 labserver 

#### Step 6: Stop the Services (Optional)

To stop and remove the containers, run:

```sh
docker-compose -f otg-compose.yaml down
```

## End-User OTG Test Case Coverage and Execution

### Supported Protocols and Test Types

| Protocol   | Test Type         | Notes                          |
|------------|-------------------|--------------------------------|
| IPv4       | Basic B2B         | `gosnappi`, `snappi` examples   |
| IS-IS      | Control Plane     | Available via `ondatra`         |
| BGP        | Control Plane     | Available via `ondatra`         |
| OSPFv2     | Conformance Suite | Modify test config for gRPC     |

### Execution Examples

- The `./example/gosnappi` and `./example/snappi` directories contain basic test cases showcasing emulated devices and IPv4 traffic flows in back-to-back scenarios.  
  For detailed execution steps, refer to the corresponding `README.md` files in each directory.

- The `./example/ondatra` directory includes standard IS-IS and BGP test cases, intended to run within the Ondatra framework using the appropriate `featureprofile` branch.  
  Please consult the `README.md` in this directory for step-by-step instructions.

- Additionally, STC OTG Service supports the execution of OTG community conformance test cases. At present, only OSPFv2 is fully supported.
For an example test case, refer to the OSPFv2 conformance test case in the OTG Conformance repository.  
**Note:** When using STC OTG Service, ensure that gRPC transport is enabled (otg_grpc_transport = True) and that the OTG host and port configurations are set correctly.

## Supported OTG APIs and GNMI Path List
Refer to `SupportedAPIsList.txt` for the latest supported OTG APIs and GNMI paths.