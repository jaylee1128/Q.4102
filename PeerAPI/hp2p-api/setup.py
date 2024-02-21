#
# The MIT License
#
# Copyright (c) 2022 ETRI
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
#
from setuptools import setup, find_packages

with open("../README.md", "r") as fh:
    long_description = fh.read()

setup(
    name='hp2p-api',
    version='1.0.5',
    description='An API for HP2P Peer',
    long_description=long_description,
    long_description_content_type="text/markdown",
    author='JAYB',
    author_email='jaylee@jayb.kr',
    url='https://github.com/ITU-T-SG11-Q8/Q.4102',
    install_requires=['grpcio', 'protobuf',], #'grpcio-tools', 
    packages=find_packages(include=["hp2papi", "hp2papi.pb2", "hp2papi.peer", "hp2papi.*"]),
    include_package_data=True,
    keywords=['hp2p', 'hp2p-api'],
    python_requires='>=3.8',
    package_data={},
    zip_safe=False,
    classifiers=[
        'Programming Language :: Python :: 3.8',
        'License :: OSI Approved :: MIT License',
    ],
)
