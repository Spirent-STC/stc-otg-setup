This example shows emulated devices and traffic flows of IPv4 capabilities.
The test results can be checked by Spirent TestCenter IQ.

How to run gosnappi OTG example case
 step1: Modify example_test.go to update otgservice ip address:port and chassis ports.
 step2: Compile to create gosnappi.test command "go test -c"
 step3: Run gosnappi example case by command "./gosnappi.test -test.v -test.run TestQuickstart"