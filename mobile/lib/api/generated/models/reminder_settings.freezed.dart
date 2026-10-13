// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'reminder_settings.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ReminderSettings {

 bool get enabled;/// Day of month for the early reminder; null disables it.
@JsonKey(includeIfNull: true, name: 'early_day') int? get earlyDay;@JsonKey(name: 'on_due_date') bool get onDueDate;/// Local time, America/Mexico_City.
@JsonKey(name: 'send_time') String get sendTime;/// Lock-screen privacy opt-in.
@JsonKey(name: 'show_amounts') bool get showAmounts;
/// Create a copy of ReminderSettings
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ReminderSettingsCopyWith<ReminderSettings> get copyWith => _$ReminderSettingsCopyWithImpl<ReminderSettings>(this as ReminderSettings, _$identity);

  /// Serializes this ReminderSettings to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ReminderSettings;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ReminderSettings&&(identical(other.enabled, _this.enabled) || other.enabled == _this.enabled)&&(identical(other.earlyDay, _this.earlyDay) || other.earlyDay == _this.earlyDay)&&(identical(other.onDueDate, _this.onDueDate) || other.onDueDate == _this.onDueDate)&&(identical(other.sendTime, _this.sendTime) || other.sendTime == _this.sendTime)&&(identical(other.showAmounts, _this.showAmounts) || other.showAmounts == _this.showAmounts));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ReminderSettings;
  return Object.hash(runtimeType,_this.enabled,_this.earlyDay,_this.onDueDate,_this.sendTime,_this.showAmounts);
}

@override
String toString() {
  final _this = this as ReminderSettings;
  return 'ReminderSettings(enabled: ${_this.enabled}, earlyDay: ${_this.earlyDay}, onDueDate: ${_this.onDueDate}, sendTime: ${_this.sendTime}, showAmounts: ${_this.showAmounts})';
}


}

/// @nodoc
abstract mixin class $ReminderSettingsCopyWith<$Res>  {
  factory $ReminderSettingsCopyWith(ReminderSettings value, $Res Function(ReminderSettings) _then) = _$ReminderSettingsCopyWithImpl;
@useResult
$Res call({
 bool enabled,@JsonKey(includeIfNull: true, name: 'early_day') int? earlyDay,@JsonKey(name: 'on_due_date') bool onDueDate,@JsonKey(name: 'send_time') String sendTime,@JsonKey(name: 'show_amounts') bool showAmounts
});




}
/// @nodoc
class _$ReminderSettingsCopyWithImpl<$Res>
    implements $ReminderSettingsCopyWith<$Res> {
  _$ReminderSettingsCopyWithImpl(this._self, this._then);

  final ReminderSettings _self;
  final $Res Function(ReminderSettings) _then;

/// Create a copy of ReminderSettings
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? enabled = null,Object? earlyDay = freezed,Object? onDueDate = null,Object? sendTime = null,Object? showAmounts = null,}) {
  return _then(ReminderSettings(
enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,earlyDay: freezed == earlyDay ? _self.earlyDay : earlyDay // ignore: cast_nullable_to_non_nullable
as int?,onDueDate: null == onDueDate ? _self.onDueDate : onDueDate // ignore: cast_nullable_to_non_nullable
as bool,sendTime: null == sendTime ? _self.sendTime : sendTime // ignore: cast_nullable_to_non_nullable
as String,showAmounts: null == showAmounts ? _self.showAmounts : showAmounts // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [ReminderSettings].
extension ReminderSettingsPatterns on ReminderSettings {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ReminderSettings value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ReminderSettings() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ReminderSettings value)  $default,){
final _that = this;
switch (_that) {
case _ReminderSettings():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ReminderSettings value)?  $default,){
final _that = this;
switch (_that) {
case _ReminderSettings() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( bool enabled, @JsonKey(includeIfNull: true, name: 'early_day')  int? earlyDay, @JsonKey(name: 'on_due_date')  bool onDueDate, @JsonKey(name: 'send_time')  String sendTime, @JsonKey(name: 'show_amounts')  bool showAmounts)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ReminderSettings() when $default != null:
return $default(_that.enabled,_that.earlyDay,_that.onDueDate,_that.sendTime,_that.showAmounts);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( bool enabled, @JsonKey(includeIfNull: true, name: 'early_day')  int? earlyDay, @JsonKey(name: 'on_due_date')  bool onDueDate, @JsonKey(name: 'send_time')  String sendTime, @JsonKey(name: 'show_amounts')  bool showAmounts)  $default,) {final _that = this;
switch (_that) {
case _ReminderSettings():
return $default(_that.enabled,_that.earlyDay,_that.onDueDate,_that.sendTime,_that.showAmounts);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( bool enabled, @JsonKey(includeIfNull: true, name: 'early_day')  int? earlyDay, @JsonKey(name: 'on_due_date')  bool onDueDate, @JsonKey(name: 'send_time')  String sendTime, @JsonKey(name: 'show_amounts')  bool showAmounts)?  $default,) {final _that = this;
switch (_that) {
case _ReminderSettings() when $default != null:
return $default(_that.enabled,_that.earlyDay,_that.onDueDate,_that.sendTime,_that.showAmounts);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ReminderSettings implements ReminderSettings {
  const _ReminderSettings({this.enabled = false, @JsonKey(includeIfNull: true, name: 'early_day') this.earlyDay = 14, @JsonKey(name: 'on_due_date') this.onDueDate = true, @JsonKey(name: 'send_time') this.sendTime = '10:00', @JsonKey(name: 'show_amounts') this.showAmounts = false});
  factory _ReminderSettings.fromJson(Map<String, dynamic> json) => _$ReminderSettingsFromJson(json);

@override@JsonKey() final  bool enabled;
/// Day of month for the early reminder; null disables it.
@override@JsonKey(includeIfNull: true, name: 'early_day') final  int? earlyDay;
@override@JsonKey(name: 'on_due_date') final  bool onDueDate;
/// Local time, America/Mexico_City.
@override@JsonKey(name: 'send_time') final  String sendTime;
/// Lock-screen privacy opt-in.
@override@JsonKey(name: 'show_amounts') final  bool showAmounts;

/// Create a copy of ReminderSettings
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ReminderSettingsCopyWith<_ReminderSettings> get copyWith => __$ReminderSettingsCopyWithImpl<_ReminderSettings>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ReminderSettingsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ReminderSettings&&(identical(other.enabled, enabled) || other.enabled == enabled)&&(identical(other.earlyDay, earlyDay) || other.earlyDay == earlyDay)&&(identical(other.onDueDate, onDueDate) || other.onDueDate == onDueDate)&&(identical(other.sendTime, sendTime) || other.sendTime == sendTime)&&(identical(other.showAmounts, showAmounts) || other.showAmounts == showAmounts));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,enabled,earlyDay,onDueDate,sendTime,showAmounts);
}

@override
String toString() {
    return 'ReminderSettings(enabled: $enabled, earlyDay: $earlyDay, onDueDate: $onDueDate, sendTime: $sendTime, showAmounts: $showAmounts)';
}


}

/// @nodoc
abstract mixin class _$ReminderSettingsCopyWith<$Res> implements $ReminderSettingsCopyWith<$Res> {
  factory _$ReminderSettingsCopyWith(_ReminderSettings value, $Res Function(_ReminderSettings) _then) = __$ReminderSettingsCopyWithImpl;
@override @useResult
$Res call({
 bool enabled,@JsonKey(includeIfNull: true, name: 'early_day') int? earlyDay,@JsonKey(name: 'on_due_date') bool onDueDate,@JsonKey(name: 'send_time') String sendTime,@JsonKey(name: 'show_amounts') bool showAmounts
});




}
/// @nodoc
class __$ReminderSettingsCopyWithImpl<$Res>
    implements _$ReminderSettingsCopyWith<$Res> {
  __$ReminderSettingsCopyWithImpl(this._self, this._then);

  final _ReminderSettings _self;
  final $Res Function(_ReminderSettings) _then;

/// Create a copy of ReminderSettings
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? enabled = null,Object? earlyDay = freezed,Object? onDueDate = null,Object? sendTime = null,Object? showAmounts = null,}) {
  return _then(_ReminderSettings(
enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as bool,earlyDay: freezed == earlyDay ? _self.earlyDay : earlyDay // ignore: cast_nullable_to_non_nullable
as int?,onDueDate: null == onDueDate ? _self.onDueDate : onDueDate // ignore: cast_nullable_to_non_nullable
as bool,sendTime: null == sendTime ? _self.sendTime : sendTime // ignore: cast_nullable_to_non_nullable
as String,showAmounts: null == showAmounts ? _self.showAmounts : showAmounts // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}

// dart format on
