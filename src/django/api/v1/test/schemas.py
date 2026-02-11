from ninja import Schema


class TestSchemaIn(Schema):
    name: str
    age: int


class TestSchemaOut(Schema):
    id: int
