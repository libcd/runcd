Simple command line utility for working with [libcd](https://github.com/libcd/libcd) and [libyaml](https://github.com/libcd/libyaml).

Compile the Yaml file to the libcd intermediate representation:

```sh
./runcd compile sample/test_1.yml > sample/test_1.json
```

Execute the intermediate representation:

```sh
./runcd exec sample/test_1.json
```

Combine the above two commands:

```sh
./runcd compile sample/test_1.yml | runcd exec
```