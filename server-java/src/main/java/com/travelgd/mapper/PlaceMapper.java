package com.travelgd.mapper;

import com.travelgd.model.Place;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface PlaceMapper {
    List<Place> findAll(@Param("tag") String tag);

    Place findById(Long id);
}
