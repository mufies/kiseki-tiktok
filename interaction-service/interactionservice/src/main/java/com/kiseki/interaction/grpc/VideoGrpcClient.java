package com.kiseki.interaction.grpc;

import org.springframework.stereotype.Service;

import com.kiseki.video.grpc.*;

import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;

@Service
@Slf4j
public class VideoGrpcClient {

  @GrpcClient("video-service")
  private VideoServiceGrpc.VideoServiceBlockingStub videoServiceStub;

  public String getVideoOwnerId(String videoId) {
    try {
      GetVideoByIdRequest request = GetVideoByIdRequest.newBuilder()
          .setVideoId(videoId)
          .build();

      GetVideoByIdResponse response = videoServiceStub.getVideoById(request);
      return response.getVideo().getUserId();
    } catch (Exception e) {
      log.error("Failed to fetch video owner for videoId: {}", videoId, e);
      return null;
    }
  }
}
