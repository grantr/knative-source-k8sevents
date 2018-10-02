require 'sinatra'
require 'rack/contrib/post_body_content_type_parser'

set :bind, '0.0.0.0'

use Rack::PostBodyContentTypeParser

post '/' do
  # Create a channel and service child
  puts "params: #{params}"
end
