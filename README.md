# onos-mlb
The xApplication for ONOS SD-RAN (µONOS Architecture) to balance load among connected cells

## Overview
The `onos-mlb` is the xApplication running over ONOS SD-RAN to balance the load among connected cells.
For the load balancing, this application adjusts neighbor cells' cell individual offset (Ocn).
If a cell becomes overloaded, this application tries to offload the cell's load to the neighbor cells that have enough capacity.
Adjusting neighbor cells' `Ocn` triggers measurement events; it triggers handover events from a cell to it's neighbor cells.
As a result, by adjusting `Ocn`, the load of overloaded cell will be offloaded to the neighbor cells.

## Limitation
As of now, `onos-mlb` application only supports the scenario where an E2 node manages only a single cell.
It does not support the E2 node that controls multiple cells.

## Algorithm description
To begin with, `onos-mlb` defines each cell's load as `the number of active UEs`, not considering other factors yet.
If a cell services the most active UEs, the `onos-mlb` application considers that the cell suffers from the heaviest load.
Then, this application defines two thresholds: (i) `overload threshold` and (ii) `target threshold`.
A cell with the load that exceeds the `overloaded threshold` is an overloaded cell.
On the other hands, a cell with the load that is less than `target threshold` has enough capacity.

With the above definition, there are two conditions.
(1) if a cell's load > `overload threshold` and its neighbor cell's load < `target threshold`, the xApplication increases `Ocn` of the neighbor cell.
(2) if a cell's load < `target threshold`, the xApplication decreases all neighbors' `Ocn`.

The increased `Ocn` makes the measurement event happening sensitively, which brings about more handover events happening to move some UEs to the neighbor cells, i.e., offloading.
On the contrary, the measurement events happen conservatively with the decreased `Ocn`; it leads to the less handover events happening to avoid neighbor cells overloaded.

The described algorithm runs periodically. By default, it is set to 10 seconds.

The `Ocn` delta value (i.e., how many the application changes Ocn value) is configurable. By default, it is set to 3 to 6.

## Interaction with other ONOS SD-RAN micro-services
Unlike other xApplications such as `onos-kpimon` and `onos-pci`, `onos-mlb` xApplication does not make a subscription with a specific service model.
In order to monitor cells, it uses `onos-uenib` and `onos-topo`.
Basically, `onos-kpimon` and `onos-pci` store the number of active UEs and cell topology to `onos-uenib`.
In addition, `onos-e2t` stores the basic cell information to `onos-topo`.
`onos-mlb` just periodically scrapes `onos-uenib` and `onos-topo`.
Then, it runs the algorithm with the scraped information as inputs.
After deciding each cell's `Ocn` values, `onos-mlb` sends the control message to the E2 node.
This control message is encoded with `RC-Pre` service model.

## Command line interface
Go to `onos-cli` and command below for each purpose.
```bash
onos-cli-594848b59d-dr6bv:~$ # to see Ocn values for each cell
onos-cli-594848b59d-dr6bv:~$ onos mlb list ocns
sCell node ID   sCell PLMN ID   sCell cell ID   sCell object ID     nCell PLMN ID   nCell cell ID   Ocn [dB]
5153            138426          1454c001        87893173159116801   138426          1454c002        0
5153            138426          1454c001        87893173159116801   138426          1454c003        6
5154            138426          1454c002        87893173159116802   138426          1454c001        0
5154            138426          1454c002        87893173159116802   138426          1454c003        6
5155            138426          1454c003        87893173159116803   138426          1454c001        -6
5155            138426          1454c003        87893173159116803   138426          1454c002        -6

onos-cli-594848b59d-dr6bv:~$ # to see mlb parameters
onos-cli-594848b59d-dr6bv:~$ onos mlb list parameters
Name                     Value
interval [sec]           10
Delta Ocn per step       3
Overload threshold [%]   100
Target threshold [%]     0
Set parameters:

onos-cli-594848b59d-dr6bv:~# to change mlb parameters
onos-cli-594848b59d-dr6bv:~$ onos mlb set parameters --interval 20
onos-cli-594848b59d-dr6bv:~$ onos mlb list parameters
Name                     Value
interval [sec]           20
Delta Ocn per step       3
Overload threshold [%]   100
Target threshold [%]     0
```