import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter/material.dart';

final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();

class ApiService {
  static final ApiService _instance = ApiService._internal();
  late Dio dio;
  final _storage = const FlutterSecureStorage();
  bool _isRefreshing = false;

  factory ApiService() {
    return _instance;
  }

  ApiService._internal() {
    dio = Dio();

    if (Platform.isAndroid) {
      dio.options.baseUrl = 'http://10.0.2.2:8080/api';
    } else {
      dio.options.baseUrl = 'http://localhost:8080/api';
    }

    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = await _storage.read(key: 'access_token');
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          return handler.next(options);
        },
        onError: (DioException e, handler) async {
          if (e.response?.statusCode == 401) {
            if (e.requestOptions.path.contains('/auth/refresh')) {
              return handler.next(e);
            }

            if (_isRefreshing) {
              return handler.next(e);
            }

            try {
              _isRefreshing = true;
              final newAccessToken = await _refreshToken();

              if (newAccessToken != null) {
                e.requestOptions.headers['Authorization'] =
                    'Bearer $newAccessToken';

                final cloneReq = await dio.fetch(e.requestOptions);
                return handler.resolve(cloneReq);
              }
            } catch (refreshError) {
              await _performLogout();
            } finally {
              _isRefreshing = false;
            }
          }
          return handler.next(e);
        },
      ),
    );
  }

  Future<String?> _refreshToken() async {
    try {
      final refreshToken = await _storage.read(key: 'refresh_token');
      if (refreshToken == null) return null;

      final response = await dio.post(
        '/auth/refresh',
        data: {'refresh_token': refreshToken},
      );

      final newAccessToken = response.data['access_token'];
      await _storage.write(key: 'access_token', value: newAccessToken);

      print("🔄 Token renovado com sucesso (Silent Refresh)!");
      return newAccessToken;
    } catch (e) {
      print("❌ Falha ao renovar token: $e");
      return null;
    }
  }

  Future<void> _performLogout() async {
    await _storage.deleteAll();
    navigatorKey.currentState?.pushNamedAndRemoveUntil(
      '/login',
      (route) => false,
    );
  }
}
