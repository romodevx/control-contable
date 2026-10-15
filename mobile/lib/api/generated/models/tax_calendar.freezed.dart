// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'tax_calendar.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxCalendar {

 FiscalYear get year;/// Periods `year-01` .. `year-12`, in order.
 List<TaxCalendarEntry> get entries;
/// Create a copy of TaxCalendar
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxCalendarCopyWith<TaxCalendar> get copyWith => _$TaxCalendarCopyWithImpl<TaxCalendar>(this as TaxCalendar, _$identity);

  /// Serializes this TaxCalendar to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxCalendar;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxCalendar&&(identical(other.year, _this.year) || other.year == _this.year)&&const DeepCollectionEquality().equals(other.entries, _this.entries));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxCalendar;
  return Object.hash(runtimeType,_this.year,const DeepCollectionEquality().hash(_this.entries));
}

@override
String toString() {
  final _this = this as TaxCalendar;
  return 'TaxCalendar(year: ${_this.year}, entries: ${_this.entries})';
}


}

/// @nodoc
abstract mixin class $TaxCalendarCopyWith<$Res>  {
  factory $TaxCalendarCopyWith(TaxCalendar value, $Res Function(TaxCalendar) _then) = _$TaxCalendarCopyWithImpl;
@useResult
$Res call({
 FiscalYear year, List<TaxCalendarEntry> entries
});




}
/// @nodoc
class _$TaxCalendarCopyWithImpl<$Res>
    implements $TaxCalendarCopyWith<$Res> {
  _$TaxCalendarCopyWithImpl(this._self, this._then);

  final TaxCalendar _self;
  final $Res Function(TaxCalendar) _then;

/// Create a copy of TaxCalendar
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? year = null,Object? entries = null,}) {
  return _then(TaxCalendar(
year: null == year ? _self.year : year // ignore: cast_nullable_to_non_nullable
as FiscalYear,entries: null == entries ? _self.entries : entries // ignore: cast_nullable_to_non_nullable
as List<TaxCalendarEntry>,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxCalendar].
extension TaxCalendarPatterns on TaxCalendar {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxCalendar value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxCalendar() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxCalendar value)  $default,){
final _that = this;
switch (_that) {
case _TaxCalendar():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxCalendar value)?  $default,){
final _that = this;
switch (_that) {
case _TaxCalendar() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( FiscalYear year,  List<TaxCalendarEntry> entries)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxCalendar() when $default != null:
return $default(_that.year,_that.entries);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( FiscalYear year,  List<TaxCalendarEntry> entries)  $default,) {final _that = this;
switch (_that) {
case _TaxCalendar():
return $default(_that.year,_that.entries);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( FiscalYear year,  List<TaxCalendarEntry> entries)?  $default,) {final _that = this;
switch (_that) {
case _TaxCalendar() when $default != null:
return $default(_that.year,_that.entries);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxCalendar implements TaxCalendar {
  const _TaxCalendar({required this.year, required  List<TaxCalendarEntry> entries}): _entries = entries;
  factory _TaxCalendar.fromJson(Map<String, dynamic> json) => _$TaxCalendarFromJson(json);

@override final  FiscalYear year;
/// Periods `year-01` .. `year-12`, in order.
 final  List<TaxCalendarEntry> _entries;
/// Periods `year-01` .. `year-12`, in order.
@override List<TaxCalendarEntry> get entries {
  if (_entries is EqualUnmodifiableListView) return _entries;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_entries);
}


/// Create a copy of TaxCalendar
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxCalendarCopyWith<_TaxCalendar> get copyWith => __$TaxCalendarCopyWithImpl<_TaxCalendar>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxCalendarToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxCalendar&&(identical(other.year, year) || other.year == year)&&const DeepCollectionEquality().equals(other.entries, _entries));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,year,const DeepCollectionEquality().hash(_entries));
}

@override
String toString() {
    return 'TaxCalendar(year: $year, entries: $entries)';
}


}

/// @nodoc
abstract mixin class _$TaxCalendarCopyWith<$Res> implements $TaxCalendarCopyWith<$Res> {
  factory _$TaxCalendarCopyWith(_TaxCalendar value, $Res Function(_TaxCalendar) _then) = __$TaxCalendarCopyWithImpl;
@override @useResult
$Res call({
 FiscalYear year, List<TaxCalendarEntry> entries
});




}
/// @nodoc
class __$TaxCalendarCopyWithImpl<$Res>
    implements _$TaxCalendarCopyWith<$Res> {
  __$TaxCalendarCopyWithImpl(this._self, this._then);

  final _TaxCalendar _self;
  final $Res Function(_TaxCalendar) _then;

/// Create a copy of TaxCalendar
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? year = null,Object? entries = null,}) {
  return _then(_TaxCalendar(
year: null == year ? _self.year : year // ignore: cast_nullable_to_non_nullable
as FiscalYear,entries: null == entries ? _self._entries : entries // ignore: cast_nullable_to_non_nullable
as List<TaxCalendarEntry>,
  ));
}


}

// dart format on
