# ThreeFold V4 3Node Guide

## Introduction

We present how to create a ThreeFold V4 DIY 3Node.

## Steps

There are 4 main steps:

- [Create an account](https://github.com/threefoldtech/tfgrid4-sdk-go/tree/development/node-registrar/tools/account)
- [Create a farm](https://github.com/threefoldtech/tfgrid4-sdk-go/tree/development/node-registrar/tools/farm)
- Create a bootstrap image using the [Zero-OS V4 Boot Generator](https://v4.bootstrap.grid.tf/)
  - Use the Farm ID created in step 2 with the proper network (e.g. main, test, qa or test)\*
- [Wipe the disks on your 3Node](https://manual.grid.tf/documentation/farmers/3node_building/4_wipe_all_disks.html)
- [Set the BIOS/UEFI](https://manual.grid.tf/documentation/farmers/3node_building/5_set_bios_uefi.html)
- Boot the node with the bootstrap image USB key

> \*Note: Until the image on the V4 Boot Generator is officially supporting Zero-OS V4 for TFGrid V4, you need to use the expert version of the boot generator and use the following kernel: `zero-os-development-zos-v4-nft-v1.11-debug-6dc6ee97d7efi`. Link: https://v4.bootstrap.grid.tf/expert
