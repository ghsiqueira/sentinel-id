class Session {
  final String id;
  final String deviceInfo;
  final String ipAddress;
  final DateTime createdAt;
  final bool isCurrent;

  Session({
    required this.id,
    required this.deviceInfo,
    required this.ipAddress,
    required this.createdAt,
    required this.isCurrent,
  });

  factory Session.fromJson(Map<String, dynamic> json) {
    return Session(
      id: json['id'],
      deviceInfo: json['device_info'] ?? 'Dispositivo Desconhecido',
      ipAddress: json['ip_address'] ?? '0.0.0.0',

      createdAt: DateTime.parse(json['created_at']),

      isCurrent: json['is_current'] ?? false,
    );
  }
}
