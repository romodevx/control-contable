// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'filed_payments_update.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$FiledPaymentsUpdate {

 List<FiledPaymentEntry> get months;
/// Create a copy of FiledPaymentsUpdate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$FiledPaymentsUpdateCopyWith<FiledPaymentsUpdate> get copyWith => _$FiledPaymentsUpdateCopyWithImpl<FiledPaymentsUpdate>(this as FiledPaymentsUpdate, _$identity);

  /// Serializes this FiledPaymentsUpdate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as FiledPaymentsUpdate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is FiledPaymentsUpdate&&const DeepCollectionEquality().equals(other.months, _this.months));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as FiledPaymentsUpdate;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.months));
}

@override
String toString() {
  final _this = this as FiledPaymentsUpdate;
  return 'FiledPaymentsUpdate(months: ${_this.months})';
}


}

/// @nodoc
abstract mixin class $FiledPaymentsUpdateCopyWith<$Res>  {
  factory $FiledPaymentsUpdateCopyWith(FiledPaymentsUpdate value, $Res Function(FiledPaymentsUpdate) _then) = _$FiledPaymentsUpdateCopyWithImpl;
@useResult
$Res call({
 List<FiledPaymentEntry> months
});




}
/// @nodoc
class _$FiledPaymentsUpdateCopyWithImpl<$Res>
    implements $FiledPaymentsUpdateCopyWith<$Res> {
  _$FiledPaymentsUpdateCopyWithImpl(this._self, this._then);

  final FiledPaymentsUpdate _self;
  final $Res Function(FiledPaymentsUpdate) _then;

/// Create a copy of FiledPaymentsUpdate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? months = null,}) {
  return _then(FiledPaymentsUpdate(
months: null == months ? _self.months : months // ignore: cast_nullable_to_non_nullable
as List<FiledPaymentEntry>,
  ));
}

}


/// Adds pattern-matching-related methods to [FiledPaymentsUpdate].
extension FiledPaymentsUpdatePatterns on FiledPaymentsUpdate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _FiledPaymentsUpdate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _FiledPaymentsUpdate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _FiledPaymentsUpdate value)  $default,){
final _that = this;
switch (_that) {
case _FiledPaymentsUpdate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _FiledPaymentsUpdate value)?  $default,){
final _that = this;
switch (_that) {
case _FiledPaymentsUpdate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<FiledPaymentEntry> months)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _FiledPaymentsUpdate() when $default != null:
return $default(_that.months);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<FiledPaymentEntry> months)  $default,) {final _that = this;
switch (_that) {
case _FiledPaymentsUpdate():
return $default(_that.months);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<FiledPaymentEntry> months)?  $default,) {final _that = this;
switch (_that) {
case _FiledPaymentsUpdate() when $default != null:
return $default(_that.months);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _FiledPaymentsUpdate implements FiledPaymentsUpdate {
  const _FiledPaymentsUpdate({required  List<FiledPaymentEntry> months}): _months = months;
  factory _FiledPaymentsUpdate.fromJson(Map<String, dynamic> json) => _$FiledPaymentsUpdateFromJson(json);

 final  List<FiledPaymentEntry> _months;
@override List<FiledPaymentEntry> get months {
  if (_months is EqualUnmodifiableListView) return _months;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_months);
}


/// Create a copy of FiledPaymentsUpdate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$FiledPaymentsUpdateCopyWith<_FiledPaymentsUpdate> get copyWith => __$FiledPaymentsUpdateCopyWithImpl<_FiledPaymentsUpdate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$FiledPaymentsUpdateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _FiledPaymentsUpdate&&const DeepCollectionEquality().equals(other.months, _months));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_months));
}

@override
String toString() {
    return 'FiledPaymentsUpdate(months: $months)';
}


}

/// @nodoc
abstract mixin class _$FiledPaymentsUpdateCopyWith<$Res> implements $FiledPaymentsUpdateCopyWith<$Res> {
  factory _$FiledPaymentsUpdateCopyWith(_FiledPaymentsUpdate value, $Res Function(_FiledPaymentsUpdate) _then) = __$FiledPaymentsUpdateCopyWithImpl;
@override @useResult
$Res call({
 List<FiledPaymentEntry> months
});




}
/// @nodoc
class __$FiledPaymentsUpdateCopyWithImpl<$Res>
    implements _$FiledPaymentsUpdateCopyWith<$Res> {
  __$FiledPaymentsUpdateCopyWithImpl(this._self, this._then);

  final _FiledPaymentsUpdate _self;
  final $Res Function(_FiledPaymentsUpdate) _then;

/// Create a copy of FiledPaymentsUpdate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? months = null,}) {
  return _then(_FiledPaymentsUpdate(
months: null == months ? _self._months : months // ignore: cast_nullable_to_non_nullable
as List<FiledPaymentEntry>,
  ));
}


}

// dart format on
