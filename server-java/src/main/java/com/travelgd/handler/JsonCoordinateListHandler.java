package com.travelgd.handler;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.travelgd.model.Route;
import org.apache.ibatis.type.BaseTypeHandler;
import org.apache.ibatis.type.JdbcType;

import java.sql.CallableStatement;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.List;

/** Maps a MySQL JSON column to List&lt;Route.Coordinate&gt;. */
public class JsonCoordinateListHandler extends BaseTypeHandler<List<Route.Coordinate>> {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    @Override
    public void setNonNullParameter(PreparedStatement ps, int i, List<Route.Coordinate> parameter, JdbcType jdbcType)
            throws SQLException {
        ps.setString(i, toJson(parameter));
    }

    @Override
    public List<Route.Coordinate> getNullableResult(ResultSet rs, String columnName) throws SQLException {
        return fromJson(rs.getString(columnName));
    }

    @Override
    public List<Route.Coordinate> getNullableResult(ResultSet rs, int columnIndex) throws SQLException {
        return fromJson(rs.getString(columnIndex));
    }

    @Override
    public List<Route.Coordinate> getNullableResult(CallableStatement cs, int columnIndex) throws SQLException {
        return fromJson(cs.getString(columnIndex));
    }

    private String toJson(List<Route.Coordinate> value) {
        try {
            return MAPPER.writeValueAsString(value);
        } catch (Exception e) {
            throw new IllegalArgumentException("Failed to serialize polyline", e);
        }
    }

    private List<Route.Coordinate> fromJson(String json) {
        if (json == null || json.isEmpty() || "null".equals(json)) {
            return List.of();
        }
        try {
            return MAPPER.readValue(json, new TypeReference<List<Route.Coordinate>>() {});
        } catch (Exception e) {
            return List.of();
        }
    }
}
