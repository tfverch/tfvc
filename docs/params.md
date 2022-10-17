# Parameters

`tfvc` accepts the following parameters. For detailed usage examples see the [Usage](basic.md) section.

| Parameter                   | Type                  | Description                                     |
| --------------------------- | --------------------- | ----------------------------------------------- |
| --include-passed, -a        | bool (default: false) | Include passed checks in console output         |
| --include-prerelease, -e    | bool (default: false) | Include prerelease versions in checks           |
| --ssh-private-key-path, -s, | string (default: "")  | Path to private key to use for SSH module calls |
| --ssh-private-key-pwd, -w   | string (default: "")  | Password for private key file if required       |

You can also see this information in your shell by running the following command.

``` zsh
tfvc --help
```