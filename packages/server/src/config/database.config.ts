import { TypeOrmModuleOptions } from '@nestjs/typeorm';
import * as path from 'path';

export const databaseConfig = (): TypeOrmModuleOptions => {
  const isProduction = process.env.NODE_ENV === 'production';

  return {
    type: 'better-sqlite3',
    database: process.env.DATABASE_PATH || path.join(process.cwd(), 'database/atoyr.sqlite'),
    nativeBinding: '../../node_modules/better-sqlite3/build/Release/better_sqlite3.node',
    entities: [path.join(process.cwd(), 'dist/**/*.entity.js')],
    synchronize: !isProduction,
    logging: !isProduction,
    dropSchema: false,
  };
};
