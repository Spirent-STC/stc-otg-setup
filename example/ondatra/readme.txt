Here we provide examples that can run under featureprofiles/Ondatra architecture. 
You can find these examples under folder featureprofiles/stcfeature. Please follow
these steps to run the examples:
Step #1. Run generate.sh
    It will clone specific commit of github.com:openconfig/featureprofiles to the 
    existing folder featureprofiles. After that we will have a featureprofiles 
    compiling enviroment;
Step #2. Check go version
    Go version should not be lower than the featureprofiles designated version at
    featureprofiles/go.mod
Step #3. Compile the example
    Suppose your selected test is /featureprofiles/stcfeature/isis/isis_basic, you can
    go to the folder and compile it by command: go test -c
    It will create a executable binary isis_basic.test
Step #4. Make sure otg and gnmi services are both started
Step #5. Modify the binding files
    Under the selected example folder, there is a script named runtest.sh. You can check it
    to know which testbed and binding files the testcase uses. You can modify the binding
    file according to your setup. You may change the otg and gnmi services' target ip:port,
    and the stc ports' position, etc.
Step #6. Run the example
    You can run runtest.sh to perform the test.
    runtest.sh with no parameter will show you the usage.
    You can perform all or one of the testcases contained in the example.
