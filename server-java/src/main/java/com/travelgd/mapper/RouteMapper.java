package com.travelgd.mapper;

import com.travelgd.model.Route;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface RouteMapper {
    int insert(Route route);

    List<Route> findByUserId(Long userId);

    Route findById(@Param("id") Long id, @Param("userId") Long userId);

    int countByUserId(Long userId);
}
