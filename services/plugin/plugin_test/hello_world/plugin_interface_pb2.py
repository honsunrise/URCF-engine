# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: plugin_interface.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
from google.protobuf import descriptor_pb2
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import any_pb2 as google_dot_protobuf_dot_any__pb2
from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='plugin_interface.proto',
  package='plugin.core',
  syntax='proto3',
  serialized_pb=_b('\n\x16plugin_interface.proto\x12\x0bplugin.core\x1a\x19google/protobuf/any.proto\x1a\x1bgoogle/protobuf/empty.proto\"E\n\x0b\x45rrorStatus\x12\x0f\n\x07message\x18\x01 \x01(\t\x12%\n\x07\x64\x65tails\x18\x02 \x03(\x0b\x32\x14.google.protobuf.Any\"\x1d\n\rDeployRequest\x12\x0c\n\x04name\x18\x01 \x01(\t2\xdb\x01\n\x0fPluginInterface\x12\x42\n\x0eInitialization\x12\x16.google.protobuf.Empty\x1a\x18.plugin.core.ErrorStatus\x12>\n\x06\x44\x65ploy\x12\x1a.plugin.core.DeployRequest\x1a\x18.plugin.core.ErrorStatus\x12\x44\n\x10UnInitialization\x12\x16.google.protobuf.Empty\x1a\x18.plugin.core.ErrorStatusb\x06proto3')
  ,
  dependencies=[google_dot_protobuf_dot_any__pb2.DESCRIPTOR,google_dot_protobuf_dot_empty__pb2.DESCRIPTOR,])




_ERRORSTATUS = _descriptor.Descriptor(
  name='ErrorStatus',
  full_name='plugin.core.ErrorStatus',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='message', full_name='plugin.core.ErrorStatus.message', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='details', full_name='plugin.core.ErrorStatus.details', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
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
  serialized_start=95,
  serialized_end=164,
)


_DEPLOYREQUEST = _descriptor.Descriptor(
  name='DeployRequest',
  full_name='plugin.core.DeployRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='plugin.core.DeployRequest.name', index=0,
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
  serialized_start=166,
  serialized_end=195,
)

_ERRORSTATUS.fields_by_name['details'].message_type = google_dot_protobuf_dot_any__pb2._ANY
DESCRIPTOR.message_types_by_name['ErrorStatus'] = _ERRORSTATUS
DESCRIPTOR.message_types_by_name['DeployRequest'] = _DEPLOYREQUEST
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

ErrorStatus = _reflection.GeneratedProtocolMessageType('ErrorStatus', (_message.Message,), dict(
  DESCRIPTOR = _ERRORSTATUS,
  __module__ = 'plugin_interface_pb2'
  # @@protoc_insertion_point(class_scope:plugin.core.ErrorStatus)
  ))
_sym_db.RegisterMessage(ErrorStatus)

DeployRequest = _reflection.GeneratedProtocolMessageType('DeployRequest', (_message.Message,), dict(
  DESCRIPTOR = _DEPLOYREQUEST,
  __module__ = 'plugin_interface_pb2'
  # @@protoc_insertion_point(class_scope:plugin.core.DeployRequest)
  ))
_sym_db.RegisterMessage(DeployRequest)



_PLUGININTERFACE = _descriptor.ServiceDescriptor(
  name='PluginInterface',
  full_name='plugin.core.PluginInterface',
  file=DESCRIPTOR,
  index=0,
  options=None,
  serialized_start=198,
  serialized_end=417,
  methods=[
  _descriptor.MethodDescriptor(
    name='Initialization',
    full_name='plugin.core.PluginInterface.Initialization',
    index=0,
    containing_service=None,
    input_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    output_type=_ERRORSTATUS,
    options=None,
  ),
  _descriptor.MethodDescriptor(
    name='Deploy',
    full_name='plugin.core.PluginInterface.Deploy',
    index=1,
    containing_service=None,
    input_type=_DEPLOYREQUEST,
    output_type=_ERRORSTATUS,
    options=None,
  ),
  _descriptor.MethodDescriptor(
    name='UnInitialization',
    full_name='plugin.core.PluginInterface.UnInitialization',
    index=2,
    containing_service=None,
    input_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    output_type=_ERRORSTATUS,
    options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_PLUGININTERFACE)

DESCRIPTOR.services_by_name['PluginInterface'] = _PLUGININTERFACE

# @@protoc_insertion_point(module_scope)
