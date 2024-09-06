# forum

This project consists in creating a web forum that allows :

    communication between users
    associating categories to posts
    liking and disliking posts and comments
    filtering posts
    upload images


## Authors

- [@Grex](https://01.kood.tech/git/Grex)
- [@timootsing](https://01.kood.tech/git/timootsing)
- [@Kairo](https://01.kood.tech/git/Kairo)
- [@AleksandrKl](https://01.kood.tech/git/AleksandrKl)
- [@anatoli.kozuhhar](https://01.kood.tech/git/anatoli.kozuhhar)

## How to run

  ```
  go run .
  ```
  Open - [Localhost](http://localhost:8080/)

## How to run in Docker

- [Install Docker](https://www.docker.com/)
- Log out and log in
- Go to project directory
- Start Docker
- Build image
  ```
  docker build -t forum .
  ```
- Start Container
  ```
  docker container run -p 8080:8080 forum
  ```
- Go to http://localhost:8080
- ctrl+C in terminal stops container

## How to delete image
- Check images
```
docker image ls
```
- Delete image
```
docker rmi -f forum
```
