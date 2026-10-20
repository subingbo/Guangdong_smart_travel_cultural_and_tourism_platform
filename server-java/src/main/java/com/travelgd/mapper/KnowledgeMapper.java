package com.travelgd.mapper;

import com.travelgd.model.Knowledge;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface KnowledgeMapper {
    List<Knowledge> searchByKeyword(@Param("keyword") String keyword, @Param("limit") int limit);
}
