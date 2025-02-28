# ThreeFold V4 3Node Guide

## Introduction

We present how to create a ThreeFold V4 DIY 3Node.

## Steps

There are 4 main steps:

- [Create an account](https://github.com/threefoldtech/tfgrid4-sdk-go/tree/development/node-registrar/tools/account)
- [Create a farm](https://github.com/threefoldtech/tfgrid4-sdk-go/tree/development/node-registrar/tools/farm)
- Create a bootstrap image using the [Zero-OS V4 Boot Generator](https://v4.bootstrap.grid.tf/)
  - Use the Farm ID created in step 2 with the proper network (e.g. main, test, qa or test)
- [Wipe the disks on your 3Node](https://manual.grid.tf/documentation/farmers/3node_building/4_wipe_all_disks.html)
- [Set the BIOS/UEFI](https://manual.grid.tf/documentation/farmers/3node_building/5_set_bios_uefi.html)
- Boot the node with the bootstrap image USB key