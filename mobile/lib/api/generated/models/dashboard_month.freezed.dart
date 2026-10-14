// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'dashboard_month.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$DashboardMonth {

@JsonKey(includeIfNull: true, name: 'income_cents') int? get incomeCents;@JsonKey(includeIfNull: true, name: 'expense_cents') int? get expenseCents;
/// Create a copy of DashboardMonth
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DashboardMonthCopyWith<DashboardMonth> get copyWith => _$DashboardMonthCopyWithImpl<DashboardMonth>(this as DashboardMonth, _$identity);

  /// Serializes this DashboardMonth to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DashboardMonth;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DashboardMonth&&(identical(other.incomeCents, _this.incomeCents) || other.incomeCents == _this.incomeCents)&&(identical(other.expenseCents, _this.expenseCents) || other.expenseCents == _this.expenseCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DashboardMonth;
  return Object.hash(runtimeType,_this.incomeCents,_this.expenseCents);
}

@override
String toString() {
  final _this = this as DashboardMonth;
  return 'DashboardMonth(incomeCents: ${_this.incomeCents}, expenseCents: ${_this.expenseCents})';
}


}

/// @nodoc
abstract mixin class $DashboardMonthCopyWith<$Res>  {
  factory $DashboardMonthCopyWith(DashboardMonth value, $Res Function(DashboardMonth) _then) = _$DashboardMonthCopyWithImpl;
@useResult
$Res call({
@JsonKey(includeIfNull: true, name: 'income_cents') int? incomeCents,@JsonKey(includeIfNull: true, name: 'expense_cents') int? expenseCents
});




}
/// @nodoc
class _$DashboardMonthCopyWithImpl<$Res>
    implements $DashboardMonthCopyWith<$Res> {
  _$DashboardMonthCopyWithImpl(this._self, this._then);

  final DashboardMonth _self;
  final $Res Function(DashboardMonth) _then;

/// Create a copy of DashboardMonth
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? incomeCents = freezed,Object? expenseCents = freezed,}) {
  return _then(DashboardMonth(
incomeCents: freezed == incomeCents ? _self.incomeCents : incomeCents // ignore: cast_nullable_to_non_nullable
as int?,expenseCents: freezed == expenseCents ? _self.expenseCents : expenseCents // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}

}


/// Adds pattern-matching-related methods to [DashboardMonth].
extension DashboardMonthPatterns on DashboardMonth {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DashboardMonth value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DashboardMonth() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DashboardMonth value)  $default,){
final _that = this;
switch (_that) {
case _DashboardMonth():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DashboardMonth value)?  $default,){
final _that = this;
switch (_that) {
case _DashboardMonth() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: true, name: 'income_cents')  int? incomeCents, @JsonKey(includeIfNull: true, name: 'expense_cents')  int? expenseCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DashboardMonth() when $default != null:
return $default(_that.incomeCents,_that.expenseCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: true, name: 'income_cents')  int? incomeCents, @JsonKey(includeIfNull: true, name: 'expense_cents')  int? expenseCents)  $default,) {final _that = this;
switch (_that) {
case _DashboardMonth():
return $default(_that.incomeCents,_that.expenseCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(includeIfNull: true, name: 'income_cents')  int? incomeCents, @JsonKey(includeIfNull: true, name: 'expense_cents')  int? expenseCents)?  $default,) {final _that = this;
switch (_that) {
case _DashboardMonth() when $default != null:
return $default(_that.incomeCents,_that.expenseCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DashboardMonth implements DashboardMonth {
  const _DashboardMonth({@JsonKey(includeIfNull: true, name: 'income_cents') required this.incomeCents, @JsonKey(includeIfNull: true, name: 'expense_cents') required this.expenseCents});
  factory _DashboardMonth.fromJson(Map<String, dynamic> json) => _$DashboardMonthFromJson(json);

@override@JsonKey(includeIfNull: true, name: 'income_cents') final  int? incomeCents;
@override@JsonKey(includeIfNull: true, name: 'expense_cents') final  int? expenseCents;

/// Create a copy of DashboardMonth
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DashboardMonthCopyWith<_DashboardMonth> get copyWith => __$DashboardMonthCopyWithImpl<_DashboardMonth>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DashboardMonthToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DashboardMonth&&(identical(other.incomeCents, incomeCents) || other.incomeCents == incomeCents)&&(identical(other.expenseCents, expenseCents) || other.expenseCents == expenseCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,incomeCents,expenseCents);
}

@override
String toString() {
    return 'DashboardMonth(incomeCents: $incomeCents, expenseCents: $expenseCents)';
}


}

/// @nodoc
abstract mixin class _$DashboardMonthCopyWith<$Res> implements $DashboardMonthCopyWith<$Res> {
  factory _$DashboardMonthCopyWith(_DashboardMonth value, $Res Function(_DashboardMonth) _then) = __$DashboardMonthCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(includeIfNull: true, name: 'income_cents') int? incomeCents,@JsonKey(includeIfNull: true, name: 'expense_cents') int? expenseCents
});




}
/// @nodoc
class __$DashboardMonthCopyWithImpl<$Res>
    implements _$DashboardMonthCopyWith<$Res> {
  __$DashboardMonthCopyWithImpl(this._self, this._then);

  final _DashboardMonth _self;
  final $Res Function(_DashboardMonth) _then;

/// Create a copy of DashboardMonth
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? incomeCents = freezed,Object? expenseCents = freezed,}) {
  return _then(_DashboardMonth(
incomeCents: freezed == incomeCents ? _self.incomeCents : incomeCents // ignore: cast_nullable_to_non_nullable
as int?,expenseCents: freezed == expenseCents ? _self.expenseCents : expenseCents // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}


}

// dart format on
