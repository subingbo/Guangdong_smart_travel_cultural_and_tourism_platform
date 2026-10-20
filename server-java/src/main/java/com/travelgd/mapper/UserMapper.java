package com.travelgd.mapper;

import com.travelgd.model.User;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

@Mapper
public interface UserMapper {
    User findByUsername(@Param("username") String username);

    User findById(@Param("id") Long id);

    int insert(User user);

    int updateProfile(User user);

    int countByUsername(@Param("username") String username);
}
