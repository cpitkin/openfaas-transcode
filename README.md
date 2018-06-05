# OpenFaaS Transcode Pipeline

This repo consists of a series of functions built for [OpenFaaS](https://github.com/openfaas/faas).

## Purpose

I have a large movie collection that I want to upload into Plex. Since modern Blue-Ray movies are quite large I wanted an easy way to transcode them down to a smaller size while still retaining quality. The series of functions in this repo allows me to do just that is a very automated fashion.

## Functions (in order)

### transcode-entrypoint

This is the entrypoint into the pipeline that is used to start he process. You can also add things like Slack integration calls here.

### transcode-worker

This is the main worker for the pipleine. The worker is called using the `/async-function` endpoint. This allows the transcoding to take hours without holding up adding new media to the backlog. Since NATS is used the backgroud we can pull media off the queue when once is done. The transcoding itself is done by a great Ruby executable found [here](https://github.com/donmelton/video_transcoding). Huge shoutout to donmelton for making a great library.

#### Steps

- Download the media from the `transcode` Minio bucket
- Transcode the media into `/tmp`
- Upload the finished file to the `complete` Minio bucket
- Delete the raw file from the local container

### transcode-move

This step moves the completed media file from the `complete` Minio bucket to another Minio server or your choice.

#### Steps

- Downlaod from the `complete` bucket
- Upload to the `media` bucket on another server

### transcode-delete

This is the cleanup step.

#### Steps

- Check for the presence of the media at the final destination
- Delete the media from the `complete` and `trancdoe` buckets on the other server