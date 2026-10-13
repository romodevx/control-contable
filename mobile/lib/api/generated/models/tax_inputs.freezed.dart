// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'tax_inputs.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxInputs {

@JsonKey(name: 'regime_code') RegimeCode get regimeCode;@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear; List<TaxInputDef> get definitions; List<TaxInputValue> get values;
/// Create a copy of TaxInputs
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxInputsCopyWith<TaxInputs> get copyWith => _$TaxInputsCopyWithImpl<TaxInputs>(this as TaxInputs, _$identity);

  /// Serializes this TaxInputs to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxInputs;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxInputs&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&const DeepCollectionEquality().equals(other.definitions, _this.definitions)&&const DeepCollectionEquality().equals(other.values, _this.values));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxInputs;
  return Object.hash(runtimeType,_this.regimeCode,_this.fiscalYear,const DeepCollectionEquality().hash(_this.definitions),const DeepCollectionEquality().hash(_this.values));
}

@override
String toString() {
  final _this = this as TaxInputs;
  return 'TaxInputs(regimeCode: ${_this.regimeCode}, fiscalYear: ${_this.fiscalYear}, definitions: ${_this.definitions}, values: ${_this.values})';
}


}

/// @nodoc
abstract mixin class $TaxInputsCopyWith<$Res>  {
  factory $TaxInputsCopyWith(TaxInputs value, $Res Function(TaxInputs) _then) = _$TaxInputsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<TaxInputDef> definitions, List<TaxInputValue> values
});




}
/// @nodoc
class _$TaxInputsCopyWithImpl<$Res>
    implements $TaxInputsCopyWith<$Res> {
  _$TaxInputsCopyWithImpl(this._self, this._then);

  final TaxInputs _self;
  final $Res Function(TaxInputs) _then;

/// Create a copy of TaxInputs
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? fiscalYear = null,Object? definitions = null,Object? values = null,}) {
  return _then(TaxInputs(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,definitions: null == definitions ? _self.definitions : definitions // ignore: cast_nullable_to_non_nullable
as List<TaxInputDef>,values: null == values ? _self.values : values // ignore: cast_nullable_to_non_nullable
as List<TaxInputValue>,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxInputs].
extension TaxInputsPatterns on TaxInputs {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxInputs value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxInputs() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxInputs value)  $default,){
final _that = this;
switch (_that) {
case _TaxInputs():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxInputs value)?  $default,){
final _that = this;
switch (_that) {
case _TaxInputs() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<TaxInputDef> definitions,  List<TaxInputValue> values)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxInputs() when $default != null:
return $default(_that.regimeCode,_that.fiscalYear,_that.definitions,_that.values);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<TaxInputDef> definitions,  List<TaxInputValue> values)  $default,) {final _that = this;
switch (_that) {
case _TaxInputs():
return $default(_that.regimeCode,_that.fiscalYear,_that.definitions,_that.values);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<TaxInputDef> definitions,  List<TaxInputValue> values)?  $default,) {final _that = this;
switch (_that) {
case _TaxInputs() when $default != null:
return $default(_that.regimeCode,_that.fiscalYear,_that.definitions,_that.values);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxInputs implements TaxInputs {
  const _TaxInputs({@JsonKey(name: 'regime_code') required this.regimeCode, @JsonKey(name: 'fiscal_year') required this.fiscalYear, required  List<TaxInputDef> definitions, required  List<TaxInputValue> values}): _definitions = definitions,_values = values;
  factory _TaxInputs.fromJson(Map<String, dynamic> json) => _$TaxInputsFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
 final  List<TaxInputDef> _definitions;
@override List<TaxInputDef> get definitions {
  if (_definitions is EqualUnmodifiableListView) return _definitions;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_definitions);
}

 final  List<TaxInputValue> _values;
@override List<TaxInputValue> get values {
  if (_values is EqualUnmodifiableListView) return _values;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_values);
}


/// Create a copy of TaxInputs
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxInputsCopyWith<_TaxInputs> get copyWith => __$TaxInputsCopyWithImpl<_TaxInputs>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxInputsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxInputs&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&const DeepCollectionEquality().equals(other.definitions, _definitions)&&const DeepCollectionEquality().equals(other.values, _values));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,fiscalYear,const DeepCollectionEquality().hash(_definitions),const DeepCollectionEquality().hash(_values));
}

@override
String toString() {
    return 'TaxInputs(regimeCode: $regimeCode, fiscalYear: $fiscalYear, definitions: $definitions, values: $values)';
}


}

/// @nodoc
abstract mixin class _$TaxInputsCopyWith<$Res> implements $TaxInputsCopyWith<$Res> {
  factory _$TaxInputsCopyWith(_TaxInputs value, $Res Function(_TaxInputs) _then) = __$TaxInputsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<TaxInputDef> definitions, List<TaxInputValue> values
});




}
/// @nodoc
class __$TaxInputsCopyWithImpl<$Res>
    implements _$TaxInputsCopyWith<$Res> {
  __$TaxInputsCopyWithImpl(this._self, this._then);

  final _TaxInputs _self;
  final $Res Function(_TaxInputs) _then;

/// Create a copy of TaxInputs
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? fiscalYear = null,Object? definitions = null,Object? values = null,}) {
  return _then(_TaxInputs(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,definitions: null == definitions ? _self._definitions : definitions // ignore: cast_nullable_to_non_nullable
as List<TaxInputDef>,values: null == values ? _self._values : values // ignore: cast_nullable_to_non_nullable
as List<TaxInputValue>,
  ));
}


}

// dart format on
