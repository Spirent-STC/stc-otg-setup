#!/usr/bin/bash

# Clone specific commit of github.com:openconfig/featureprofiles to the existing featureprofiles folder;
# After that we will have a featureprofiles compiling enviroment;
# We can enter featureprofiles/stcfeature to compile & run the examples.

COMMIT=a21d1577a2c8396cfbf4411e2c2bfcb5b8999ad4

cd featureprofiles

mv stcfeature stcfeature~
git init
#git remote add origin git@github.com:openconfig/featureprofiles.git
git remote add origin https://github.com/openconfig/featureprofiles.git 
git pull origin $COMMIT --allow-unrelated-histories
mv stcfeature~ stcfeature

cd -

#done!
