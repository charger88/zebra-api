# Zebra Sharing System

Zebra Sharing System is open-source software for secured data exchange. You can run it and as docker containers or as regular software. 

## Components

* Zebra API - Go app, which handles server-side part of the functionality. https://github.com/charger88/zebra-api
* Zebra Client - HTML single page app which may be served by web server or may be opened as HTML file from local disc. https://github.com/charger88/zebra-client
* Zebra docker images - possible way to run Zebra API and Zebra Client. https://github.com/charger88/zebra-docker

# Zebra API

## Requirements

* Go language
* Go packages:
  * github.com/go-yaml/yaml
  * github.com/mediocregopher/radix.v2/redis
  * golang.org/x/crypto/bcrypt
* Redis server

## How to run it

Make sure that you meet all requirement, made all required changes in the configuration and run this application with command like:
```
go run *.go
```
Please, don't expose the app into internet without _nginx_ or other web server with enabled SSL in front of it.

Also this web server may serve web client files. Example of _nginx_ configuration: https://github.com/charger88/zebra-docker/blob/master/zebra-client/default.conf

## Configuration

There are two way to change Zebra API configuration:

* Create file `config/config.yaml` and override in it values of `config/default.yaml` 
* Define environmental variables (`my-var` in config file transforms to `ZEBRA_MY_VAR` in environment)

File `config/config.yaml` has more priority than environmental variables.  

### Config file overview

#### Redis configuration

* __redis-host__ (`string`, `"127.0.0.1"`) - redis host
* __redis-port__ (`integer`, `6379`) - redis port
* __redis-password__ (`string`, `""`) - redis password
* __redis-database__ (`integer`, `0`) - redis database number
* __redis-key-prefix__ (`string`, `""`) - redis key prefix

#### HTTP configuration

* __http-interface__ (`string`, `""`) - interface for listening by API application. Provide IP or hostname. Leave empty value for listening on all interfaces 
* __http-port__ (`integer`, `8080`) - port for listening by API application
* __trusted-proxy__ (`string[]`, `- "127.0.0.1/32"`) - list of trusted proxies (your web server IP)

#### Key generation policy

* __minimal-key-length__ (`integer`, `4`) - minimal length of key for shared text
* __expected-stripes-per-hour__ (`integer`, `1000`) - expected number of shared text in one hour
* __appropriate-chance-to-guess__ (`integer`, `1000000000`) - the greater this value is, the longer key will be generated 

#### Rate limiting

* __allowed-bad-attempts__ (`integer`, `5`) - number of allowed failed attempts to retrieve text in one minute
* __allowed-shares-period__ (`integer`, `60`) - rate limit period for text sharing (seconds)
* __allowed-shares-number-in-period__ (`integer`, `5`) - rate limit for text sharing (number in `allowed-shares-period` seconds) 

#### Security configuration

* __max-expiration-time__ (`integer`, `86400`) - maximal text's expiration time (in seconds)
* __max-text-length__ (`integer`, `50000`) - total text length limit in bytes (so it is not accurate after encryption)
* __password-policy__ (`string`, `"allowed"`) - possible values are:
    * `allowed` - password for shared text is optional
    * `required` - password for shared text is required
    * `disabled` - password for shared text not allowed
* __encryption-password-policy__ (`string`, `"allowed"`) - option for client application, possible values are:
    * `allowed` - (different) encryption password for shared text is optional
    * `required` - (different) encryption password for shared text is required
    * `disabled` - (different) encryption password for shared text not allowed
* __require-api-key__ (`boolean`, `false`) - require `X-Api-Key`
* __require-api-key-for-post-only__ (`boolean`, `true`) - require `X-Api-Key` for text sharing only (`require-api-key` should be `true`)
* __allowed-api-keys__ (`string[]`) - list of appropriate values of `X-Api-Key` header 

#### Configuration for client

* __public-name__ (`string`, `"Zebra Sharing Service"`) - name of the instance
* __public-color__ (`string`, `"#425766"`) - color of header in client
* __public-url__ (`string`, `"https://127.0.0.1/"`) - URL of your web server which serves client and proxy API
* __public-email__ (`string`, `""`) - email of current instance administrator 

#### System configuration

* __version__ (`string`, `"1.0.0"`) - API version. You don't need to override it.
* __config-reload-time__ (`string`, `60`) - time in seconds for configs reload (this option, as well as `http-interface` and `http-port` will not being updated without app restart)
* __extended-logs__ (`boolean`, `false`) - log all events from Zebra API. Don't enable if you are not absolutely sure about protection of log files.

## API Overview

This is JSON REST API.

API allows header `X-Api-Key` for API Key (optional, see _require-api-key_, _require-api-key-for-post-only_ and _allowed-api-keys_ configurations). 

All API routes also support `OPTIONS` HTTP method.

### /

#### GET

##### Response (json)

* __routes__ - list of REST API resources (routes) and allowed methods
    * __%route name%__ - list of allowed HTTP methods

### /ping

#### GET

##### Response (json)

* __timestamp__ - UNIX timestamp current

### /config

#### GET

##### Response (json)

* __version__ - API version (from config __version__)
* __name__ - instance name (from config __public-name__)
* __url__ - URL of web server (from config __public-url__)
* __email__ - administrator's email (from config __public-email__)
* __color__ - color for client (from config __public-color__)
* __max-expiration-time__ - max expiration time (from config __max-expiration-time__)
* __max-text-length__  - total text length limit in bytes (so it is not accurate after encryption)
* __encryption-password-policy__ - client-side encryption password policy (from config __password-policy__)
* __password-policy__ - password policy (from config __password-policy__)
* __require-api-key__ - require API key configuration (from config __require-api-key__)
* __require-api-key-for-post-only__ - require API key for POST only configuration (from config __require-api-key-for-post-only__)

### /stripe

#### GET

##### Request query string

* __key__ - key
* __password__ - password _(optional)_
* __check-key__ - special key which allows to ignore rate limiting. It is being generated when text is deleted _(optional)_

##### Response (json)

* __key__ - key
* __data__ - share data
* __expiration__ - expiration timestamp
* __burn__ - `true` if text will be deleted after this opening (actually it is already deleted) 

#### POST

##### Request (json)

* __data__ - sharing data
* __burn__ - `true` to delete after the first opening
* __expiration__ - expiration in seconds
* __mode__ - key generation mode (`uppercase-lowercase-digits`, `uppercase-digits`, `uppercase`, `digits`)
* __encrypted-with-client-side-password__ - confirmation of encryption on client-side with client-side password
* __password__ -  password _(optional)_

##### Response (json)

* __key__ - key
* __expiration__ - expiration timestamp
* __owner-key__ - owner's key (required for deleting)

#### DELETE

##### Request query string

* __key__ - key
* __owner-key__ - owner's key

##### Response (json)

* __success__ - boolean value of deletion's success
* __check-key__ - string key which allows to ignore rate limiting for attempt to load the text (so you can check text's non-existence without rate limit)