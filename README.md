# pgg
Post from the Get-Go

### Synopsis
`pgg [-m GET] [-e foo] [-c config-file] [-f "param_name=./foo.txt"] "http://example.com/{{my_var}}"`

### Description
Pgg - Post from the Get-Go.
Pgg is a client you can use to send http requests while taking advantage of custom environments set in a configuration file.

### Options
-m, -method
   Specify the request method.

-e, -env
    Specify the environment to use.

-c, -cfg
    Specify an alternative config file.

-f, -file
    Specify the file to upload.

-h
    Prints the help message.

--help
    Prints options details.

## Files
When starting pgg looks for configuration files in the following order:

You can specify a custom path using the -c option.
	1. ~/.config/pgg/config
	2. ~/.pgg/config

The config file must be written in TOML style.
The variables defined in the environments can be accessed by enclosing their name in double curly brackets.

For example running:
 pgg -e foo 'http://example.com/{{foo}}/{{name}}'

will result in a GET request to
 http://example.com/bar/pgg.

Sample configuration

```TOML
# pgg config file

# Set the default environment.
default_env = "foo"
[env]
   # Add a custom environment named "foo".
   [env.foo]
   scheme = "http://"
   vars = [
       "foo=bar",
       "name=pgg"
   ]

   # Add a custom environment named "bar".
   [env.bar]
   scheme = "ftp://"
   vars = [
       "base_url=www.example.com"
   ]
```

## Environment
`$HOME`
This variable is used to read the config file $HOME/.config/pgg/config
