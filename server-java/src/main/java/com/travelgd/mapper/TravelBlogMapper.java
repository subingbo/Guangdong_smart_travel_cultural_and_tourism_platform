package com.travelgd.mapper;

import com.travelgd.model.TravelBlog;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface TravelBlogMapper {
    List<TravelBlog> findAll(@Param("viewerId") Long viewerId);

    TravelBlog findById(@Param("id") Long id, @Param("viewerId") Long viewerId);

    int insert(TravelBlog blog);

    int incrementLikes(Long id);

    int countByUserId(Long userId);

    int addLike(@Param("userId") Long userId, @Param("blogId") Long blogId);
}
