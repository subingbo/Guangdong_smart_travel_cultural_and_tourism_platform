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

/** Maps a MySQL JSON column to List&lt;Route.Waypoint&gt;. */
public class JsonWaypointListHandler extends BaseTypeHandler<List<Route.Waypoint>> {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    @Override
    public void setNonNullParameter(PreparedStatement ps, int i, List<Route.Waypoint> parameter, JdbcType jdbcType)
            throws SQLException {
        ps.setString(i, toJson(parameter));
    }

    @Override
    public List<Route.Waypoint> getNullableResult(ResultSet rs, String columnName) throws SQLException {
        return fromJson(rs.getString(columnName));
    }

    @Override
    public List<Route.Waypoint> getNullableResult(ResultSet rs, int columnIndex) throws SQLException {
        return fromJson(rs.getString(columnIndex));
    }

    @Override
    public List<Route.Waypoint> getNullableResult(CallableStatement cs, int columnIndex) throws SQLException {
        return fromJson(cs.getString(columnIndex));
    }

    private String toJson(List<Route.Waypoint> value) {
        try {
            return MAPPER.writeValueAsString(value);
        } catch (Exception e) {
            throw new IllegalArgumentException("Failed to serialize waypoints", e);
        }
    }

    private List<Route.Waypoint> fromJson(String json) {
        if (json == null || json.isEmpty() || "null".equals(json)) {
            return List.of();
        }
        try {
            return MAPPER.readValue(json, new TypeReference<List<Route.Waypoint>>() {});
        } catch (Exception e) {
            return List.of();
        }
    }
}
