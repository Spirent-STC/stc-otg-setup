# Spirent OTG Service Deployment Guide

The Spirent OTG Service is only available for use on Linux PC/workstations and depends on the installation of the Spirent ReST API application. You can pre-install either the Spirent TestCenter application or Spirent Labserver as the ReST API application.

- Recommend system requirements for OTG Service application:
        • 2 core CPU
        • 2 GB RAM
        • 30 GB disk (SSD or better)

- Recommend using STC version later than 5.52.
- The OTG Service can be installed on the same machine as the ReST API application.
- If using **Spirent AION Licensing**, refer to the *Spirent TestCenter with Spirent AION Licensing Quick Start Guide (DOC12187)* for further details.


## Deploying OTG Service Directly
  - Suitable for standalone installations on a single machine.
  - Requires manually running the installation script and configuring services.
  - Best for cases where Docker is not used or when installing OTG service alongside a ReST API application.

### Step 1: Obtain the Installation File

Download the executable file:

- **File Name:** `otgservice.V[x.xx].sh`

### Step 2: Install the OTG Service

Run the installer to extract the files into a new folder with the same name as the package:

```bash
./otgservice.V[x.xx].sh
```

### Step 3: Start OTG and gNMI Services

Navigate to the extracted folder and start the services using the default `otg.conf` file:

```bash
cd otgservice.V[x.xx]
./otgctl --start
```

### Step 4: Configure the STC ReST Server IP

Set the ReST server IP address:

```bash
./otgctl --restserver 1.2.3.4:80
```

### Step 5: Access Logs

Logs are automatically saved in:

```bash
./otgservice.V[x.xx]/Logs/
```


- To view all available commands, run:

```bash
./otgctl --help
```
---

## Deploying OTG Service as a Docker Container
  - Automates deployment using Docker containers for both OTG and Labserver services.
  - Supports scalability, allowing multiple OTG instances to be created dynamically.
  - Ideal for test labs or cloud environments where services need to be managed and deployed efficiently.
### Environment
Install Docker engine and Docker Compose on any flavor of Linux VM.

- **Docker version:** 27.3.1
- **Docker Compose version:** 1.29.2

### Step 1: Clone the Repository

Download the required setup files by cloning the repository:

```bash
git clone https://github.com/SpirentOrion/stc-otg-setup
cd stc-otg-setup
```

**Required files:**
- `otg-compose.yaml`: Main Docker Compose YAML file
- `.env`: Environment variables defined for Docker Compose file
- `Dockerfile`: Dockerfile to build OTG services
- `entrypoint.sh`: Shell script used to start OTG services (GNMI and OTG)
- `otg-multi-compose.yaml`: Docker Compose file to start multiple instances of OTG services
- `otgservice.V[x.xx]`: Spirent OTG Service application

### Step 2: Load the Labserver Docker Image

If a specific Labserver version is needed, download and load it into Docker:

```bash
docker load -i labserver-5.49.2816.tar.xz
```

### Step 3: Update Configuration Files

Modify the `.env` file to set the required **environment variables** and specify the **OTG service binary file**.

### Step 4: Deploy OTG & Labserver Services

Run the Docker Compose file to start the services:

```bash
docker-compose -f otg-compose.yaml up -d
```

### Step 5: Deploy Multiple OTG Instances (Optional)

To run multiple instances of the OTG service, use the following command:

```bash
docker-compose -f otg-multi-compose.yaml up --scale otg=2 -d
```

- The `--scale otg=` option allows you to create multiple OTG/gNMI service instances.
- **Dynamic Port Ranges:**
  - **OTG Service:** 48153–48200
  - **gNMI Service:** 49153–49200

### Step 6: Stop the Services (Optional)

To stop and remove the containers, run:

```bash
docker-compose -f otg-compose.yaml down
```


### End-user OTG Test Case Script Execution Examples

Both **`./example/gosnappi`** and **`./example/snappi`** applied cases demonstrate basic emulated devices and traffic flows with IPv4 capabilities in back-to-back scenarios.
**`./example/ondatra`** includes typical IS-IS and BGP test cases, designed to run within the Ondatra framework using the specified featureprofile branch.

For detailed instructions on running these examples, refer to the **README guide** in each example case folder.


### Supported OTG APIs and GNMI Path List

Refer to each **release note** for the latest supported OTG APIs and GNMI path list.
