import { TypeOrmModuleOptions } from '@nestjs/typeorm';
import * as path from 'path';

export const databaseConfig = (): TypeOrmModuleOptions => {
  const isProduction = process.env.NODE_ENV === 'production';

  return {
    type: 'sqlite',
    database: process.env.DATABASE_PATH || path.join(process.cwd(), 'database/atoyr.sqlite'),
    entities: [path.join(process.cwd(), 'dist/**/*.entity.js')],
    synchronize: !isProduction,
    logging: !isProduction,
    dropSchema: false,
  };
};
