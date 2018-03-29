# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: command.proto

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


DESCRIPTOR = _descriptor.FileDescriptor(
  name='command.proto',
  package='plugin.protocol',
  syntax='proto3',
  serialized_pb=_b('\n\rcommand.proto\x12\x0fplugin.protocol\x1a\x19google/protobuf/any.proto\"E\n\x0b\x45rrorStatus\x12\x0f\n\x07message\x18\x01 \x01(\t\x12%\n\x07\x64\x65tails\x18\x02 \x03(\x0b\x32\x14.google.protobuf.Any\"D\n\x0e\x43ommandRequest\x12\x0c\n\x04name\x18\x01 \x01(\t\x12$\n\x06params\x18\x02 \x03(\x0b\x32\x14.google.protobuf.Any\"(\n\x12\x43ommandHelprequest\x12\x12\n\nsubcommand\x18\x01 \x01(\t\"\x1f\n\x0f\x43ommandHelpResp\x12\x0c\n\x04help\x18\x01 \x01(\t2\xae\x01\n\x10\x43ommandInterface\x12H\n\x07\x43ommand\x12\x1f.plugin.protocol.CommandRequest\x1a\x1c.plugin.protocol.ErrorStatus\x12P\n\x07GetHelp\x12#.plugin.protocol.CommandHelprequest\x1a .plugin.protocol.CommandHelpRespb\x06proto3')
  ,
  dependencies=[google_dot_protobuf_dot_any__pb2.DESCRIPTOR,])




_ERRORSTATUS = _descriptor.Descriptor(
  name='ErrorStatus',
  full_name='plugin.protocol.ErrorStatus',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='message', full_name='plugin.protocol.ErrorStatus.message', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='details', full_name='plugin.protocol.ErrorStatus.details', index=1,
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
  serialized_start=61,
  serialized_end=130,
)


_COMMANDREQUEST = _descriptor.Descriptor(
  name='CommandRequest',
  full_name='plugin.protocol.CommandRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='plugin.protocol.CommandRequest.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='params', full_name='plugin.protocol.CommandRequest.params', index=1,
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
  serialized_start=132,
  serialized_end=200,
)


_COMMANDHELPREQUEST = _descriptor.Descriptor(
  name='CommandHelprequest',
  full_name='plugin.protocol.CommandHelprequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='subcommand', full_name='plugin.protocol.CommandHelprequest.subcommand', index=0,
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
  serialized_start=202,
  serialized_end=242,
)


_COMMANDHELPRESP = _descriptor.Descriptor(
  name='CommandHelpResp',
  full_name='plugin.protocol.CommandHelpResp',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='help', full_name='plugin.protocol.CommandHelpResp.help', index=0,
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
  serialized_start=244,
  serialized_end=275,
)

_ERRORSTATUS.fields_by_name['details'].message_type = google_dot_protobuf_dot_any__pb2._ANY
_COMMANDREQUEST.fields_by_name['params'].message_type = google_dot_protobuf_dot_any__pb2._ANY
DESCRIPTOR.message_types_by_name['ErrorStatus'] = _ERRORSTATUS
DESCRIPTOR.message_types_by_name['CommandRequest'] = _COMMANDREQUEST
DESCRIPTOR.message_types_by_name['CommandHelprequest'] = _COMMANDHELPREQUEST
DESCRIPTOR.message_types_by_name['CommandHelpResp'] = _COMMANDHELPRESP
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

ErrorStatus = _reflection.GeneratedProtocolMessageType('ErrorStatus', (_message.Message,), dict(
  DESCRIPTOR = _ERRORSTATUS,
  __module__ = 'command_pb2'
  # @@protoc_insertion_point(class_scope:plugin.protocol.ErrorStatus)
  ))
_sym_db.RegisterMessage(ErrorStatus)

CommandRequest = _reflection.GeneratedProtocolMessageType('CommandRequest', (_message.Message,), dict(
  DESCRIPTOR = _COMMANDREQUEST,
  __module__ = 'command_pb2'
  # @@protoc_insertion_point(class_scope:plugin.protocol.CommandRequest)
  ))
_sym_db.RegisterMessage(CommandRequest)

CommandHelprequest = _reflection.GeneratedProtocolMessageType('CommandHelprequest', (_message.Message,), dict(
  DESCRIPTOR = _COMMANDHELPREQUEST,
  __module__ = 'command_pb2'
  # @@protoc_insertion_point(class_scope:plugin.protocol.CommandHelprequest)
  ))
_sym_db.RegisterMessage(CommandHelprequest)

CommandHelpResp = _reflection.GeneratedProtocolMessageType('CommandHelpResp', (_message.Message,), dict(
  DESCRIPTOR = _COMMANDHELPRESP,
  __module__ = 'command_pb2'
  # @@protoc_insertion_point(class_scope:plugin.protocol.CommandHelpResp)
  ))
_sym_db.RegisterMessage(CommandHelpResp)



_COMMANDINTERFACE = _descriptor.ServiceDescriptor(
  name='CommandInterface',
  full_name='plugin.protocol.CommandInterface',
  file=DESCRIPTOR,
  index=0,
  options=None,
  serialized_start=278,
  serialized_end=452,
  methods=[
  _descriptor.MethodDescriptor(
    name='Command',
    full_name='plugin.protocol.CommandInterface.Command',
    index=0,
    containing_service=None,
    input_type=_COMMANDREQUEST,
    output_type=_ERRORSTATUS,
    options=None,
  ),
  _descriptor.MethodDescriptor(
    name='GetHelp',
    full_name='plugin.protocol.CommandInterface.GetHelp',
    index=1,
    containing_service=None,
    input_type=_COMMANDHELPREQUEST,
    output_type=_COMMANDHELPRESP,
    options=None,
  ),
])
_sym_db.RegisterServiceDescriptor(_COMMANDINTERFACE)

DESCRIPTOR.services_by_name['CommandInterface'] = _COMMANDINTERFACE

# @@protoc_insertion_point(module_scope)
