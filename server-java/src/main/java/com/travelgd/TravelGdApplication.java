package com.travelgd;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
@MapperScan("com.travelgd.mapper")
public class TravelGdApplication {
    public static void main(String[] args) {
        SpringApplication.run(TravelGdApplication.class, args);
    }
}
