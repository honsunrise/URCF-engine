# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: plugin_interface.proto

import sys

_b = sys.version_info[0] < 3 and (lambda x: x) or (lambda x: x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database

# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()

DESCRIPTOR = _descriptor.FileDescriptor(
    name='plugin_interface.proto',
    package='proto',
    syntax='proto3',
    serialized_pb=_b(
        '\n\x16plugin_interface.proto\x12\x05proto\".\n\x0b\x45rrorStatus\x12\x0f\n\x05\x65rror\x18\x02 \x01(\tH\x00\x42\x0e\n\x0coptional_err\"\x1d\n\rDeployRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\"\x07\n\x05\x45mpty2\xaf\x01\n\x0fPluginInterface\x12\x32\n\x0eInitialization\x12\x0c.proto.Empty\x1a\x12.proto.ErrorStatus\x12\x32\n\x06\x44\x65ploy\x12\x14.proto.DeployRequest\x1a\x12.proto.ErrorStatus\x12\x34\n\x10UnInitialization\x12\x0c.proto.Empty\x1a\x12.proto.ErrorStatusb\x06proto3')
)

_ERRORSTATUS = _descriptor.Descriptor(
    name='ErrorStatus',
    full_name='proto.ErrorStatus',
    filename=None,
    file=DESCRIPTOR,
    containing_type=None,
    fields=[
        _descriptor.FieldDescriptor(
            name='error', full_name='proto.ErrorStatus.error', index=0,
            number=2, type=9, cpp_type=9, label=1,
            has_default_value=False, default_value=_b("").decode('utf-8'),
            message_type=None, enum_type=None, containing_type=None,
            is_extension=False, extension_scope=None,
            options=None, file=DESCRIPTOR),
    ],
    extensions=[
    ],
    nested_types=[],
    enum_types=[
    ],
    options=None,
    is_extendable=False,
    syntax='proto3',
    extension_ranges=[],
    oneofs=[
        _descriptor.OneofDescriptor(
            name='optional_err', full_name='proto.ErrorStatus.optional_err',
            index=0, containing_type=None, fields=[]),
    ],
    serialized_start=33,
    serialized_end=79,
)

_DEPLOYREQUEST = _descriptor.Descriptor(
    name='DeployRequest',
    full_name='proto.DeployRequest',
    filename=None,
    file=DESCRIPTOR,
    containing_type=None,
    fields=[
        _descriptor.FieldDescriptor(
            name='name', full_name='proto.DeployRequest.name', index=0,
            number=1, type=9, cpp_type=9, label=1,
            has_default_value=False, default_value=_b("").decode('utf-8'),
            message_type=None, enum_type=None, containing_type=None,
            is_extension=False, extension_scope=None,
            options=None, file=DESCRIPTOR),
    ],
    extensions=[
    ],
    nested_types=[],
    enum_types=[
    ],
    options=None,
    is_extendable=False,
    syntax='proto3',
    extension_ranges=[],
    oneofs=[
    ],
    serialized_start=81,
    serialized_end=110,
)

_EMPTY = _descriptor.Descriptor(
    name='Empty',
    full_name='proto.Empty',
    filename=None,
    file=DESCRIPTOR,
    containing_type=None,
    fields=[
    ],
    extensions=[
    ],
    nested_types=[],
    enum_types=[
    ],
    options=None,
    is_extendable=False,
    syntax='proto3',
    extension_ranges=[],
    oneofs=[
    ],
    serialized_start=112,
    serialized_end=119,
)

_ERRORSTATUS.oneofs_by_name['optional_err'].fields.append(
    _ERRORSTATUS.fields_by_name['error'])
_ERRORSTATUS.fields_by_name['error'].containing_oneof = _ERRORSTATUS.oneofs_by_name['optional_err']
DESCRIPTOR.message_types_by_name['ErrorStatus'] = _ERRORSTATUS
DESCRIPTOR.message_types_by_name['DeployRequest'] = _DEPLOYREQUEST
DESCRIPTOR.message_types_by_name['Empty'] = _EMPTY
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

ErrorStatus = _reflection.GeneratedProtocolMessageType('ErrorStatus', (_message.Message,), dict(
    DESCRIPTOR=_ERRORSTATUS,
    __module__='plugin_interface_pb2'
    # @@protoc_insertion_point(class_scope:proto.ErrorStatus)
))
_sym_db.RegisterMessage(ErrorStatus)

DeployRequest = _reflection.GeneratedProtocolMessageType('DeployRequest', (_message.Message,), dict(
    DESCRIPTOR=_DEPLOYREQUEST,
    __module__='plugin_interface_pb2'
    # @@protoc_insertion_point(class_scope:proto.DeployRequest)
))
_sym_db.RegisterMessage(DeployRequest)

Empty = _reflection.GeneratedProtocolMessageType('Empty', (_message.Message,), dict(
    DESCRIPTOR=_EMPTY,
    __module__='plugin_interface_pb2'
    # @@protoc_insertion_point(class_scope:proto.Empty)
))
_sym_db.RegisterMessage(Empty)

_PLUGININTERFACE = _descriptor.ServiceDescriptor(
    name='PluginInterface',
    full_name='proto.PluginInterface',
    file=DESCRIPTOR,
    index=0,
    options=None,
    serialized_start=122,
    serialized_end=297,
    methods=[
        _descriptor.MethodDescriptor(
            name='Initialization',
            full_name='proto.PluginInterface.Initialization',
            index=0,
            containing_service=None,
            input_type=_EMPTY,
            output_type=_ERRORSTATUS,
            options=None,
        ),
        _descriptor.MethodDescriptor(
            name='Deploy',
            full_name='proto.PluginInterface.Deploy',
            index=1,
            containing_service=None,
            input_type=_DEPLOYREQUEST,
            output_type=_ERRORSTATUS,
            options=None,
        ),
        _descriptor.MethodDescriptor(
            name='UnInitialization',
            full_name='proto.PluginInterface.UnInitialization',
            index=2,
            containing_service=None,
            input_type=_EMPTY,
            output_type=_ERRORSTATUS,
            options=None,
        ),
    ])
_sym_db.RegisterServiceDescriptor(_PLUGININTERFACE)

DESCRIPTOR.services_by_name['PluginInterface'] = _PLUGININTERFACE

# @@protoc_insertion_point(module_scope)
