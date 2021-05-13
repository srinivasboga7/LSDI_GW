# LSDI

LSDI (lightweight and scalable DAG based blockchain for verifying IoT data integrity).


## Table of contents
- Background
- Usage
    - Prerequisites
    - Build
    - API
- Development

## Background

This repository has a implementation of LSDI gateway node and discovery node. LSDI is a DAG based blockchain system inspired from IoTA to improve the scalabilty of using blockchain in IoT environments. To find more about the system please read https://ieeexplore.ieee.org/abstract/document/9334000/.

## Usage

### Prerequisites

- golang (version > 1.08) installed
- Discovery node up and running
- At least one storage node up and running. For storage node code check [this repository](https://github.com/sumanthcherupally/LSDI_SN)
- provide the address of the discovery node in bootstrapNodes.txt file

### Build

You can build the node by running :
> go build main.go 


### API


## Development 

