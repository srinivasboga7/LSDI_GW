# LSDI

LSDI (lightweight and scalable DAG based blockchain for verifying IoT data integrity).


## Background

This repository has a implementation of LSDI gateway node and discovery node. LSDI is a DAG based blockchain system inspired from IoTA to improve the scalabilty of using blockchain in IoT environments. To find more about the system please read https://ieeexplore.ieee.org/abstract/document/9334000/.

## Usage

### Prerequisites

- golang (version > 1.08) installed
- Discovery node up and running
- At least one storage node up and running. For storage node code check [this repository](https://github.com/sumanthcherupally/LSDI_SN)
- provide the address of the discovery node in bootstrapNodes.txt file

### Building the Source 

You can build the node by running :
> go build main.go 

A sample DockerFile is provided in the repository to run the node in a docker container.

### Running Discovery Node
The code for discovery node is provided in the DiscoveryNode directory in the repo. Build the discoveryService.go file by running
> go build discoveryService.go

To run the discovery node we need to provide port number, max nodes in a shard as command line arguments
> ./discoveryService port_number max_nodes


### API

To generate transactions and access information related to the node an API is provided. REST API is hosted on port 8989.

### Generating Transactions

To generate a transaction you need to provide a hash value (Integrity Proof of the Data) to be stored in the transaction.

SampleUrl    | http://192.168.0.2:8989/api/ (hash value)
---          | ---
Request Type | HTTP GET Request
Response     | Transaction ID

### Transaction Input Rate

Transaction Input Rate is the number of valid transactions per second in the node that are processed and attached to the DAG. 

SampleUrl    | http://192.168.0.2:8989/TxInputRate/
---          | ---
Request Type | HTTP GET Request
Response     | JSON object containing Transaction Input Rate

In a similar way there is an API provided to access Memory Usage, CPU usage etc.

To know more about the other functionality api provides read client.go 

## Pseudo Gateway Application

The repo also contains a simple IoT gateway functionality integrated with LSDI. The gateway module provides a REST API to accept data from the sensors and forward it to the cloud database (database address to be provided in the config.json file). 

SampleUrl   | http://192.168.0.2:7000/postData 
---         | ---
RequestType | HTTP POST Request
Response    | HTTP OK 


## Future Development

To contribute to this project, fork the repo and start creating a pull request for the repository maintainers to review and merge into the main branch. 

To get the documentation of all the modules use godoc feature, for example :
> godoc -http=:6060
 
The above commands generates doumentation for the entire repo, you can access the documentation by going to localhost:6060 on your browser.





