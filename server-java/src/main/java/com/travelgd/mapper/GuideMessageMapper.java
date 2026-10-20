package com.travelgd.mapper;

import com.travelgd.model.GuideSession;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface GuideMessageMapper {
    int insert(GuideSession.GuideMessage message);

    List<GuideSession.GuideMessage> findBySessionId(Long sessionId);

    int countBySessionId(Long sessionId);
}
