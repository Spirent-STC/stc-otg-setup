This example shows emulated devices and traffic flows of IPv4 capabilities.
The test results can be checked by Spirent TestCenter IQ.

How to run snappi OTG example case
 step1: Modify otg-flows.py to update otgservice ip address:port and chassis ports.
 step2: Excute any one of following command to run example case
         python3 otg-flows.py -m port #For port matrices
         python3 otg-flows.py -m flow #For flow matrices
