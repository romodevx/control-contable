// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_breakdown_line.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualBreakdownLine {

 int get seq;@JsonKey(includeIfNull: true, name: 'regime_code') RegimeCode? get regimeCode;/// Human-readable step (Spanish).
 String get concept; AnnualBreakdownLineOperation get operation;@JsonKey(name: 'amount_cents') MoneyCents get amountCents;@JsonKey(includeIfNull: true) DecimalRate? get rate;@JsonKey(name: 'legal_basis') List<Citation> get legalBasis; AnnualLineSource get source;@JsonKey(name: 'needs_cpa_review') bool get needsCpaReview;
/// Create a copy of AnnualBreakdownLine
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualBreakdownLineCopyWith<AnnualBreakdownLine> get copyWith => _$AnnualBreakdownLineCopyWithImpl<AnnualBreakdownLine>(this as AnnualBreakdownLine, _$identity);

  /// Serializes this AnnualBreakdownLine to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualBreakdownLine;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualBreakdownLine&&(identical(other.seq, _this.seq) || other.seq == _this.seq)&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.concept, _this.concept) || other.concept == _this.concept)&&(identical(other.operation, _this.operation) || other.operation == _this.operation)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.rate, _this.rate) || other.rate == _this.rate)&&const DeepCollectionEquality().equals(other.legalBasis, _this.legalBasis)&&(identical(other.source, _this.source) || other.source == _this.source)&&(identical(other.needsCpaReview, _this.needsCpaReview) || other.needsCpaReview == _this.needsCpaReview));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualBreakdownLine;
  return Object.hash(runtimeType,_this.seq,_this.regimeCode,_this.concept,_this.operation,_this.amountCents,_this.rate,const DeepCollectionEquality().hash(_this.legalBasis),_this.source,_this.needsCpaReview);
}

@override
String toString() {
  final _this = this as AnnualBreakdownLine;
  return 'AnnualBreakdownLine(seq: ${_this.seq}, regimeCode: ${_this.regimeCode}, concept: ${_this.concept}, operation: ${_this.operation}, amountCents: ${_this.amountCents}, rate: ${_this.rate}, legalBasis: ${_this.legalBasis}, source: ${_this.source}, needsCpaReview: ${_this.needsCpaReview})';
}


}

/// @nodoc
abstract mixin class $AnnualBreakdownLineCopyWith<$Res>  {
  factory $AnnualBreakdownLineCopyWith(AnnualBreakdownLine value, $Res Function(AnnualBreakdownLine) _then) = _$AnnualBreakdownLineCopyWithImpl;
@useResult
$Res call({
 int seq,@JsonKey(includeIfNull: true, name: 'regime_code') RegimeCode? regimeCode, String concept, AnnualBreakdownLineOperation operation,@JsonKey(name: 'amount_cents') MoneyCents amountCents,@JsonKey(includeIfNull: true) DecimalRate? rate,@JsonKey(name: 'legal_basis') List<Citation> legalBasis, AnnualLineSource source,@JsonKey(name: 'needs_cpa_review') bool needsCpaReview
});




}
/// @nodoc
class _$AnnualBreakdownLineCopyWithImpl<$Res>
    implements $AnnualBreakdownLineCopyWith<$Res> {
  _$AnnualBreakdownLineCopyWithImpl(this._self, this._then);

  final AnnualBreakdownLine _self;
  final $Res Function(AnnualBreakdownLine) _then;

/// Create a copy of AnnualBreakdownLine
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? seq = null,Object? regimeCode = freezed,Object? concept = null,Object? operation = null,Object? amountCents = null,Object? rate = freezed,Object? legalBasis = null,Object? source = null,Object? needsCpaReview = null,}) {
  return _then(AnnualBreakdownLine(
seq: null == seq ? _self.seq : seq // ignore: cast_nullable_to_non_nullable
as int,regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,concept: null == concept ? _self.concept : concept // ignore: cast_nullable_to_non_nullable
as String,operation: null == operation ? _self.operation : operation // ignore: cast_nullable_to_non_nullable
as AnnualBreakdownLineOperation,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,rate: freezed == rate ? _self.rate : rate // ignore: cast_nullable_to_non_nullable
as DecimalRate?,legalBasis: null == legalBasis ? _self.legalBasis : legalBasis // ignore: cast_nullable_to_non_nullable
as List<Citation>,source: null == source ? _self.source : source // ignore: cast_nullable_to_non_nullable
as AnnualLineSource,needsCpaReview: null == needsCpaReview ? _self.needsCpaReview : needsCpaReview // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [AnnualBreakdownLine].
extension AnnualBreakdownLinePatterns on AnnualBreakdownLine {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualBreakdownLine value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualBreakdownLine() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualBreakdownLine value)  $default,){
final _that = this;
switch (_that) {
case _AnnualBreakdownLine():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualBreakdownLine value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualBreakdownLine() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int seq, @JsonKey(includeIfNull: true, name: 'regime_code')  RegimeCode? regimeCode,  String concept,  AnnualBreakdownLineOperation operation, @JsonKey(name: 'amount_cents')  MoneyCents amountCents, @JsonKey(includeIfNull: true)  DecimalRate? rate, @JsonKey(name: 'legal_basis')  List<Citation> legalBasis,  AnnualLineSource source, @JsonKey(name: 'needs_cpa_review')  bool needsCpaReview)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualBreakdownLine() when $default != null:
return $default(_that.seq,_that.regimeCode,_that.concept,_that.operation,_that.amountCents,_that.rate,_that.legalBasis,_that.source,_that.needsCpaReview);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int seq, @JsonKey(includeIfNull: true, name: 'regime_code')  RegimeCode? regimeCode,  String concept,  AnnualBreakdownLineOperation operation, @JsonKey(name: 'amount_cents')  MoneyCents amountCents, @JsonKey(includeIfNull: true)  DecimalRate? rate, @JsonKey(name: 'legal_basis')  List<Citation> legalBasis,  AnnualLineSource source, @JsonKey(name: 'needs_cpa_review')  bool needsCpaReview)  $default,) {final _that = this;
switch (_that) {
case _AnnualBreakdownLine():
return $default(_that.seq,_that.regimeCode,_that.concept,_that.operation,_that.amountCents,_that.rate,_that.legalBasis,_that.source,_that.needsCpaReview);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int seq, @JsonKey(includeIfNull: true, name: 'regime_code')  RegimeCode? regimeCode,  String concept,  AnnualBreakdownLineOperation operation, @JsonKey(name: 'amount_cents')  MoneyCents amountCents, @JsonKey(includeIfNull: true)  DecimalRate? rate, @JsonKey(name: 'legal_basis')  List<Citation> legalBasis,  AnnualLineSource source, @JsonKey(name: 'needs_cpa_review')  bool needsCpaReview)?  $default,) {final _that = this;
switch (_that) {
case _AnnualBreakdownLine() when $default != null:
return $default(_that.seq,_that.regimeCode,_that.concept,_that.operation,_that.amountCents,_that.rate,_that.legalBasis,_that.source,_that.needsCpaReview);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualBreakdownLine implements AnnualBreakdownLine {
  const _AnnualBreakdownLine({required this.seq, @JsonKey(includeIfNull: true, name: 'regime_code') required this.regimeCode, required this.concept, required this.operation, @JsonKey(name: 'amount_cents') required this.amountCents, @JsonKey(includeIfNull: true) required this.rate, @JsonKey(name: 'legal_basis') required  List<Citation> legalBasis, required this.source, @JsonKey(name: 'needs_cpa_review') required this.needsCpaReview}): _legalBasis = legalBasis;
  factory _AnnualBreakdownLine.fromJson(Map<String, dynamic> json) => _$AnnualBreakdownLineFromJson(json);

@override final  int seq;
@override@JsonKey(includeIfNull: true, name: 'regime_code') final  RegimeCode? regimeCode;
/// Human-readable step (Spanish).
@override final  String concept;
@override final  AnnualBreakdownLineOperation operation;
@override@JsonKey(name: 'amount_cents') final  MoneyCents amountCents;
@override@JsonKey(includeIfNull: true) final  DecimalRate? rate;
 final  List<Citation> _legalBasis;
@override@JsonKey(name: 'legal_basis') List<Citation> get legalBasis {
  if (_legalBasis is EqualUnmodifiableListView) return _legalBasis;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_legalBasis);
}

@override final  AnnualLineSource source;
@override@JsonKey(name: 'needs_cpa_review') final  bool needsCpaReview;

/// Create a copy of AnnualBreakdownLine
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualBreakdownLineCopyWith<_AnnualBreakdownLine> get copyWith => __$AnnualBreakdownLineCopyWithImpl<_AnnualBreakdownLine>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualBreakdownLineToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualBreakdownLine&&(identical(other.seq, seq) || other.seq == seq)&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.concept, concept) || other.concept == concept)&&(identical(other.operation, operation) || other.operation == operation)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.rate, rate) || other.rate == rate)&&const DeepCollectionEquality().equals(other.legalBasis, _legalBasis)&&(identical(other.source, source) || other.source == source)&&(identical(other.needsCpaReview, needsCpaReview) || other.needsCpaReview == needsCpaReview));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,seq,regimeCode,concept,operation,amountCents,rate,const DeepCollectionEquality().hash(_legalBasis),source,needsCpaReview);
}

@override
String toString() {
    return 'AnnualBreakdownLine(seq: $seq, regimeCode: $regimeCode, concept: $concept, operation: $operation, amountCents: $amountCents, rate: $rate, legalBasis: $legalBasis, source: $source, needsCpaReview: $needsCpaReview)';
}


}

/// @nodoc
abstract mixin class _$AnnualBreakdownLineCopyWith<$Res> implements $AnnualBreakdownLineCopyWith<$Res> {
  factory _$AnnualBreakdownLineCopyWith(_AnnualBreakdownLine value, $Res Function(_AnnualBreakdownLine) _then) = __$AnnualBreakdownLineCopyWithImpl;
@override @useResult
$Res call({
 int seq,@JsonKey(includeIfNull: true, name: 'regime_code') RegimeCode? regimeCode, String concept, AnnualBreakdownLineOperation operation,@JsonKey(name: 'amount_cents') MoneyCents amountCents,@JsonKey(includeIfNull: true) DecimalRate? rate,@JsonKey(name: 'legal_basis') List<Citation> legalBasis, AnnualLineSource source,@JsonKey(name: 'needs_cpa_review') bool needsCpaReview
});




}
/// @nodoc
class __$AnnualBreakdownLineCopyWithImpl<$Res>
    implements _$AnnualBreakdownLineCopyWith<$Res> {
  __$AnnualBreakdownLineCopyWithImpl(this._self, this._then);

  final _AnnualBreakdownLine _self;
  final $Res Function(_AnnualBreakdownLine) _then;

/// Create a copy of AnnualBreakdownLine
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? seq = null,Object? regimeCode = freezed,Object? concept = null,Object? operation = null,Object? amountCents = null,Object? rate = freezed,Object? legalBasis = null,Object? source = null,Object? needsCpaReview = null,}) {
  return _then(_AnnualBreakdownLine(
seq: null == seq ? _self.seq : seq // ignore: cast_nullable_to_non_nullable
as int,regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,concept: null == concept ? _self.concept : concept // ignore: cast_nullable_to_non_nullable
as String,operation: null == operation ? _self.operation : operation // ignore: cast_nullable_to_non_nullable
as AnnualBreakdownLineOperation,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,rate: freezed == rate ? _self.rate : rate // ignore: cast_nullable_to_non_nullable
as DecimalRate?,legalBasis: null == legalBasis ? _self._legalBasis : legalBasis // ignore: cast_nullable_to_non_nullable
as List<Citation>,source: null == source ? _self.source : source // ignore: cast_nullable_to_non_nullable
as AnnualLineSource,needsCpaReview: null == needsCpaReview ? _self.needsCpaReview : needsCpaReview // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}

// dart format on
