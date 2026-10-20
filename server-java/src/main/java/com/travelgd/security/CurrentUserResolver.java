package com.travelgd.security;

import com.travelgd.common.ApiException;
import org.springframework.core.MethodParameter;
import org.springframework.stereotype.Component;
import org.springframework.web.bind.support.WebDataBinderFactory;
import org.springframework.web.context.request.NativeWebRequest;
import org.springframework.web.method.support.HandlerMethodArgumentResolver;
import org.springframework.web.method.support.ModelAndViewContainer;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

import java.util.List;

/** Resolves @CurrentUser parameters from the Authorization header. */
@Component
public class CurrentUserResolver implements HandlerMethodArgumentResolver, WebMvcConfigurer {

    private final JwtService jwtService;

    public CurrentUserResolver(JwtService jwtService) {
        this.jwtService = jwtService;
    }

    @Override
    public boolean supportsParameter(MethodParameter parameter) {
        return parameter.hasParameterAnnotation(CurrentUser.class);
    }

    @Override
    public Object resolveArgument(MethodParameter parameter, ModelAndViewContainer mavContainer,
                                  NativeWebRequest webRequest, WebDataBinderFactory binderFactory) {
        String auth = webRequest.getHeader("Authorization");
        if (auth == null || !auth.startsWith("Bearer ")) {
            throw ApiException.unauthorized("Missing bearer token");
        }
        Long userId = jwtService.parseUserId(auth.substring(7));
        if (userId == null) {
            throw ApiException.unauthorized("Invalid or expired token");
        }
        return userId;
    }

    @Override
    public void addArgumentResolvers(List<HandlerMethodArgumentResolver> resolvers) {
        resolvers.add(this);
    }
}
