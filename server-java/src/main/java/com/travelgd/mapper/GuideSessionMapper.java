package com.travelgd.mapper;

import com.travelgd.model.GuideSession;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

@Mapper
public interface GuideSessionMapper {
    int insert(GuideSession session);

    GuideSession findById(@Param("id") Long id, @Param("userId") Long userId);

    int updateTitle(GuideSession session);
}
