# Docker Builder

**Docker Builder** is a `Dockerfile` templating tool. It allows one to turn a `Dockerfile.template`:

```
FROM %%BASE_IMAGE%%

%%INSTALL_COMMON_PACKAGES%%

RUN apk add apache2 %%PHP_PACKAGES%%

COPY %%_COMMON/scripts/mount-nfs.sh%% /mount-nfs.sh
COPY src/run-server.sh /run-server.sh
RUN chmod +x /run-server.sh

ENTRYPOINT ["/run-server.sh"]
```

... into a `Dockerfile`:

```
FROM alpine:3.9

RUN apk add bash
RUN apk add shadow
RUN apk add wget

RUN apk add apache2 php7

COPY scripts/mount-nfs.sh%% /mount-nfs.sh
COPY src/run-server.sh /run-server.sh
RUN chmod +x /run-server.sh

ENTRYPOINT ["/run-server.sh"]
```

## Download and Install

Download **Docker Builder** from [the release page](https://github.com/qiushihe/docker-builder/releases).

Put the extracted `docker-builder` CLI binary somewhere on your `PATH`, or not since **Docker Builder** doesn't care where it's being run from nor does it care where all the directories are.

## How to Use Docker Builder

To use **Docker Builder**, first setup a container image directory structure as follow:

* `/path/to/my-image/`
  * `src/`
    * `entry.sh`

      ```
      #!/usr/bin/env bash
      # This is the entry script, so do whatever you want in here.
      echo "Hallo!"
      while true; do sleep 5; done
      ```

  * `Dockerfile.template`

    ```
    FROM %%BASE_IMAGE%%
    COPY src/entry.sh /entry.sh
    RUN chmod +x /run-server.sh
    ENTRYPOINT ["/run-server.sh"]
    ```

  * `Dockerfile.variables`

    ```
    BASE_IMAGE: alpine:3.9
    ```

... then run the `docker-builder` CLI tool:

```
$ docker-builder -src /path/to/my-image -- -t my-image
```

Running the above command would generate a `_docker-build` directory that looks like:

* `/path/to/my-image/_docker-build`
  * `src/`
    * `entry.sh`
  * `Dockerfile`

    ```
    FROM alpine:3.9
    COPY src/entry.sh /entry.sh
    RUN chmod +x /run-server.sh
    ENTRYPOINT ["/run-server.sh"]
    ```

... and then invoke `docker build -t my-image` to build the image.

### Reference Block Values

Unlike one line string values which are stored in `Dockerfile.variables`, block values are stored in `Dockerfile.variables.d` directories.

Consider this example:

* `/path/to/my-image/`
  * `src/`
    * `entry.sh`
  * `Dockerfile.template`

    ```
    FROM %%BASE_IMAGE%%
    %%INSTALL_PACKAGES%%
    COPY src/entry.sh /entry.sh
    RUN chmod +x /run-server.sh
    ENTRYPOINT ["/run-server.sh"]
    ```

  * `Dockerfile.variables`

    ```
    BASE_IMAGE: alpine:3.9
    ```

  * `Dockerfile.variables.d/`
    * INSTALL_PACKAGES
      
      ```
      RUN apk add bash wget \
        zip unzip \
        shadow
      ```

The above setup would generate this:

* `/path/to/my-image/_docker-build`
  * `src/`
    * `entry.sh`
  * `Dockerfile`

    ```
    FROM alpine:3.9
    RUN apk add bash wget \
        zip unzip \
        shadow
    COPY src/entry.sh /entry.sh
    RUN chmod +x /run-server.sh
    ENTRYPOINT ["/run-server.sh"]
    ```

### Reference External Values

Other than `Dockerfile.variables` and `Dockerfile.variables.d` within the image's directory itself, **Docker Builder** can also be told to take any number of other directories into consideration.

Consider this example:

* `/path/to/somewhere/else/`
  * `configs/`
    * `httpd.conf`

      ```
      ... some apache2 httpd configuration ...
      ```
  * `Dockerfile.variables`

    ```
    APACHE_PORTS: 80 443
    ```
  
  * `Dockerfile.variables.d`
    * `APACHE_PACKAGES`

      ```
      RUN apk add apache2 apache2-proxy
      ```

* `/path/to/my-image/`
  * `src/`
    * `entry.sh`
  * `Dockerfile.template`

    ```
    FROM %%BASE_IMAGE%%
    %%INSTALL_PACKAGES%%
    %%APACHE_PACKAGES%%
    COPY %%_COMMON/configs/httpd.conf%% /etc/apache2/httpd.conf
    COPY src/entry.sh /entry.sh
    EXPOSE %%APACHE_PORTS%%
    RUN chmod +x /run-server.sh
    ENTRYPOINT ["/run-server.sh"]
    ```

  * `Dockerfile.variables`

    ```
    BASE_IMAGE: alpine:3.9
    ```

  * `Dockerfile.variables.d/`
    * INSTALL_PACKAGES
      
      ```
      RUN apk add bash wget \
        zip unzip \
        shadow
      ```

Running this command:

```
$ docker-build -src /path/to/my-image -lib /path/to/somewhere/else -- -t my-image
```

... would generate:

* `/path/to/my-image/_docker-build`
  * `src/`
    * `entry.sh`
  * `configs/`
    * `httpd.conf`
  * `Dockerfile`

    ```
    FROM alpine:3.9
    RUN apk add bash wget \
        zip unzip \
        shadow
    RUN apk add apache2 apache2-proxy
    COPY configs/httpd.conf /etc/apache2/httpd.conf
    COPY src/entry.sh /entry.sh
    EXPOSE 80 443
    RUN chmod +x /run-server.sh
    ENTRYPOINT ["/run-server.sh"]
    ```

And you can specify multiple `-lib /PATH/TO/LIB` options.
