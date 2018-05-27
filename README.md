# Useful tips

If you need to run redis in Docker for test purposes on VM and want to expose port - do this:

```
sudo docker run -d -p 6379:6379 redis
```