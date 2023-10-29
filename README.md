# urlfetcher cli tool
CLI tool that make async http requests to 3rd-party services urls.

1. build it by running make build
```
make build 
```
2. make your config by running executable with --prepare flag and setting workers
```
cd build && build/ctrl_{platform} --prepare 
```
3. fill urls in file delimited by \n
4. run it - output in stdout
```
cd build && /ctrl_{platform} --filepath=path to file in 3.
```