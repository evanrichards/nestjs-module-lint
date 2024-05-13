import { Module } from "@nestjs/common";
import { SomeImport } from "./some-import";
import { OtherName as SomeOtherImport } from "src/some-other-import";
import ThirdModule from "@src/third-module";
import { ThisService, ThatService } 'src/service'

@Module({
  imports: [SomeImport, SomeOtherImport],
  exports: [ThisService],
  providers: [],
})
export class TestModule {}

@Module({
  imports: [ThirdModule],
  exports: [ThatService],
  providers: [],
})
export class OtherModule {}
