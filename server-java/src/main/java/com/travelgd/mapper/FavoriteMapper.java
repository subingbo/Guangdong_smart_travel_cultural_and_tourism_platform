package com.travelgd.mapper;

import com.travelgd.model.Place;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface FavoriteMapper {
    List<Place> findPlacesByUserId(Long userId);

    int insert(@Param("userId") Long userId, @Param("placeId") Long placeId);

    int delete(@Param("userId") Long userId, @Param("placeId") Long placeId);

    boolean exists(@Param("userId") Long userId, @Param("placeId") Long placeId);

    int countByUserId(Long userId);
}
