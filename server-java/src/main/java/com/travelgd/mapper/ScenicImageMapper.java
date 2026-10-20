package com.travelgd.mapper;

import com.travelgd.model.ScenicImage;
import org.apache.ibatis.annotations.Mapper;

import java.util.List;

@Mapper
public interface ScenicImageMapper {
    List<ScenicImage> findAll();

    ScenicImage findById(Long id);
}
